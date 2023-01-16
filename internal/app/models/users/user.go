package users

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

var patterns = map[string]string{
	"email": `\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`,
}

type ID uint64
type Email string
type User struct {
	ID        ID     `db:"id" json:"id"`
	Email     Email  `db:"email" json:"email"`
	PwdHash   string `db:"pwd_hash" json:"pwd_hash"`
	Status    string `db:"status" json:"status"`
	Confirmed bool   `db:"confirmed" json:"confirmed"`
}

func NewUserID() ID {
	return ID(uuid.New().ID())
}

func NewUser(email Email, pwdhash string, status string) User {
	return User{
		ID:        NewUserID(),
		Email:     email,
		PwdHash:   pwdhash,
		Status:    status,
		Confirmed: false,
	}
}

func ValidateEmail(email Email) bool {
	check, _ := regexp.MatchString(patterns["email"], string(email))
	return check
}

type Session struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

func (usr *User) CheckCerds(email Email, pwdhash string) bool {
	if usr.Email != email {
		return false
	}

	return doPasswordsMatch(string(usr.PwdHash), string(pwdhash))
}

// TODO: перенести в auth сервис
func doPasswordsMatch(hashedPassword, currPassword string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword), []byte(currPassword))
	return err == nil
}
