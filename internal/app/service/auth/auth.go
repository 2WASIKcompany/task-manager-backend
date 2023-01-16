package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
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
	InvalidEmail          = errors.New("invalid email")
	UserAlreadyExist      = errors.New("user already exist")
	IncorrectCreds        = errors.New("incorrect creds")
	NotFoundEmail         = errors.New("not found email")
	NotFoundRestoreUIDErr = errors.New("not found confirm uid")
	NotFoundEmailErr      = errors.New("email not found")
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

	CashRefreshToken(users.ID, string, time.Duration) error
	GetUserIDByRefreshToken(string) (string, error)
	DeleteSession(string) error
	CreateRestoreUID(users.Email, users.ID, string) error
	GetUserIDByRestoreUID(string) (users.ID, error)
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
	if users.ValidateEmail(email) {
		return InvalidEmail
	}

	if _, err := s.repository.GetUserByEmail(ctx, email); err != nil {
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

func (s *Service) Auth(ctx context.Context, password string, email users.Email) (users.Session, error) {
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("Auth: Cant get user: %v", err)
		return users.Session{}, err
	}

	credsCorrect := user.CheckCerds(users.Email(email), password)

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

func (s *Service) ChangePassword(ctx context.Context, restoreUID, newPassword string) error {
	userID, err := s.repository.GetUserIDByRestoreUID(restoreUID)
	if err != nil {
		return NotFoundRestoreUIDErr
	}

	saltPass, err := salt(newPassword)
	if err != nil {
		return err
	}

	return s.repository.ChangePasswordByUserID(ctx, userID, saltPass)
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

func (s *Service) ConfirmationUser(ctx context.Context, refresh string) error {
	strUserID, err := s.repository.GetUserIDByRefreshToken(refresh)
	if err != nil {
		return NotFoundRestoreUIDErr
	}
	userID, _ := strconv.Atoi(strUserID)
	return s.repository.ConfirmUser(ctx, users.ID(userID))
}

func (s *Service) SendRestorePasswordMail(ctx context.Context, email users.Email) error {
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return NotFoundEmail
	}

	uid := strconv.Itoa(int(uuid.New().ID()))
	err = s.repository.CreateRestoreUID(email, user.ID, uid)
	if err != nil {
		return err
	}

	s.sender.SendMail(string(email), fmt.Sprintf(changePassMsg, string(email), uid))

	return nil
}
