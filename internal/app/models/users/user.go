package users

import "github.com/google/uuid"

type User struct {
	ID        uint64 `db:"id" json:"id"`
	Email     string `db:"email" json:"email"`
	PwdHash   string `db:"pwd_hash" json:"pwd_hash"`
	Status    string `db:"status" json:"status"`
	Confirmed bool   `db:"confirmed" json:"confirmed"`
}

func NewUserID() uint64 {
	return uint64(uuid.New().ID())
}

func NewUser(email string, pwdhash string, status string) User {
	return User{
		ID:      NewUserID(),
		Email:   email,
		PwdHash: pwdhash,
		Status:  status,
	}
}
