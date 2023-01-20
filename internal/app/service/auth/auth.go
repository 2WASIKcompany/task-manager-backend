package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"regexp"
	"strconv"
	"task-manager-backend/internal/app/models/users"
	"task-manager-backend/internal/app/service/mail"
	"time"
)

const (
	SessionTime    = 86400
	RefreshKeyTime = 43200
)

var (
	InvalidData      = errors.New("Invalid data")
	UserAlreadyExist = errors.New("This email was registered before")
	IncorrectCreds   = errors.New("Incorrect login or password")
	NotFoundEmail    = errors.New("This email is not registered")
	InvalidRefresh   = errors.New("Invalid token")
	NonConfirmed     = errors.New("Registration has not been confirmed")
	SamePassword     = errors.New("Same password")
)

type AuthData struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type Repository interface {
	CreateUser(context.Context, users.User) error
	GetUserByEmail(context.Context, users.Email) (users.User, error)
	ConfirmUser(context.Context, users.ID) error
	ChangePasswordByUserID(context.Context, users.ID, string) error
	GetUserByUserID(context.Context, users.ID) (users.User, error)

	CashRefreshToken(users.ID, string, time.Duration) error
	GetUserIDByRefreshToken(string) (string, error)
	DeleteSession(string) error
	CreateRestoreRefresh(users.Email, users.ID, string) error
	GetUserIDByRestoreRefresh(string) (users.ID, error)
}

func NewService(repository Repository, jwt *Manager, sender *mail.Sender) *Service {
	return &Service{
		repository: repository,
		jwt:        jwt,
		sender:     sender,
	}
}

type Service struct {
	repository Repository
	jwt        *Manager
	sender     *mail.Sender
}

func (s *Service) Register(ctx context.Context, password string, email users.Email) error {
	if !users.ValidateEmail(email) || !validatePass(password) {
		return InvalidData
	}

	if user, _ := s.repository.GetUserByEmail(ctx, email); user.Email == email {
		return UserAlreadyExist
	}

	saltPass, err := salt(password)
	if err != nil {
		return err
	}

	user := users.NewUser(users.Email(email), saltPass, "simple")
	err = s.repository.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	refresh := s.jwt.CreateRefreshToken()
	err = s.repository.CashRefreshToken(user.ID, refresh, RefreshKeyTime*time.Second)
	if err != nil {
		return err
	}

	err = s.sender.SendMail(string(email), fmt.Sprintf(confimEmailMsg, email, refresh))

	return err
}

func salt(pass string) (string, error) {
	saltPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	return string(saltPass), err
}

func (s *Service) Logout(ctx context.Context, refresh string) error {
	return s.repository.DeleteSession(refresh)
}

func (s *Service) Auth(ctx context.Context, password string, email users.Email) (users.Session, error) {
	if !users.ValidateEmail(email) || !validatePass(password) {
		return users.Session{}, InvalidData
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if user.Email == "" || err != nil {
		log.Printf("Auth: Cant get user: %v", err)
		return users.Session{}, IncorrectCreds
	}

	if !user.Confirmed {
		return users.Session{}, NonConfirmed
	}

	credsCorrect := user.CheckCerds(email, password)

	if !credsCorrect {
		return users.Session{}, IncorrectCreds
	}

	token, err := s.jwt.CreateToken(user.ID)
	if err != nil {
		log.Printf("Auth: Cant CreateToken: %v", err)
	}

	refresh := s.jwt.CreateRefreshToken()

	session := users.Session{token, refresh}
	err = s.repository.CashRefreshToken(user.ID, refresh, SessionTime*time.Second)
	if err != nil {
		log.Printf("Auth: Cant Add Session: %v", err)
	}

	return session, err
}

func (s *Service) ChangePassword(ctx context.Context, restoreRefresh, newPassword string) (users.Session, error) {
	if !validatePass(newPassword) {
		return users.Session{}, InvalidData
	}

	userID, err := s.repository.GetUserIDByRestoreRefresh(restoreRefresh)
	if err != nil {
		return users.Session{}, InvalidRefresh
	}

	user, err := s.repository.GetUserByUserID(ctx, userID)
	if users.DoPasswordsMatch(user.PwdHash, newPassword) {
		return users.Session{}, SamePassword
	}

	saltPass, err := salt(newPassword)
	if err != nil {
		return users.Session{}, err
	}

	token, _ := s.jwt.CreateToken(userID)
	refresh := s.jwt.CreateRefreshToken()

	return users.Session{Token: token, Refresh: refresh}, s.repository.ChangePasswordByUserID(ctx, userID, saltPass)
}

func validatePass(pass string) bool {
	if lenPass := len(pass); lenPass < 6 || lenPass > 20 {
		return false
	}
	if check, _ := regexp.MatchString("^[\\x20-\\x7E]+$", pass); !check {
		return false
	}
	if check, _ := regexp.MatchString("[0-9]", pass); !check {
		return false
	}
	if check, _ := regexp.MatchString("[A-Za-z]", pass); !check {
		return false
	}
	return true
}

func (s *Service) UnmarshalToken(token string) (users.ID, error) {
	userID, err := s.jwt.GetIDFromToken(token)
	return users.ID(userID), err
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (users.Session, error) {
	userID, err := s.repository.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		log.Printf("RefreshToken: Cant Get Session: %v", err)
		return users.Session{}, err
	}

	uid, _ := strconv.Atoi(userID)
	token, err := s.jwt.CreateToken(users.ID(uid))
	if err != nil {
		log.Printf("Auth: Cant CreateToken: %v", err)
		return users.Session{}, err
	}

	newRefreshToken := s.jwt.CreateRefreshToken()
	if err = s.repository.DeleteSession(refreshToken); err != nil {
		return users.Session{}, err
	}
	err = s.repository.CashRefreshToken(users.ID(uid), newRefreshToken, SessionTime*time.Second)
	if err != nil {
		log.Printf("Auth: Cant Add Session: %v", err)
		return users.Session{}, err
	}

	return users.Session{Token: token, Refresh: newRefreshToken}, nil
}

func (s *Service) ConfirmationUser(ctx context.Context, restoreRefresh string) (users.Session, error) {
	strUserID, err := s.repository.GetUserIDByRefreshToken(restoreRefresh)
	if err != nil {
		return users.Session{}, InvalidRefresh
	}
	s.repository.DeleteSession(restoreRefresh)
	userID, _ := strconv.Atoi(strUserID)

	token, _ := s.jwt.CreateToken(users.ID(userID))
	refresh := s.jwt.CreateRefreshToken()

	return users.Session{Token: token, Refresh: refresh}, s.repository.ConfirmUser(ctx, users.ID(userID))
}

func (s *Service) SendRestorePasswordMail(ctx context.Context, email users.Email) error {
	if !users.ValidateEmail(email) {
		return InvalidData
	}
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return NotFoundEmail
	}

	refresh := s.jwt.CreateRefreshToken()
	err = s.repository.CreateRestoreRefresh(email, user.ID, refresh)
	if err != nil {
		return err
	}

	s.sender.SendMail(string(email), fmt.Sprintf(changePassMsg, string(email), refresh))

	return nil
}
