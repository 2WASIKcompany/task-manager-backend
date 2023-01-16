package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"strconv"
	"task-manager-backend/internal/app/config"
	"task-manager-backend/internal/app/models/users"
	"time"
)

type InvalidTokenError struct {
	message string
}

func (e *InvalidTokenError) Error() string {
	return e.message
}

var (
	ErrInvalidToken   = &InvalidTokenError{"invalid token"}
	ErrExpiredToken   = &InvalidTokenError{"expired token"}
	ErrInvalidSubject = &InvalidTokenError{"invalid subject"}
)

type Manager struct {
	Secret     string
	Expiration time.Duration
}

func NewManger(cfg config.ServiceConfiguration) *Manager {
	return &Manager{
		Secret:     cfg.SignKey,
		Expiration: cfg.TokenTTL,
	}
}

func (m *Manager) CreateToken(uid users.ID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(m.Expiration).Unix(),
		Subject:   strconv.FormatInt(int64(uid), 10),
	})

	return token.SignedString([]byte(m.Secret))
}

func (m *Manager) CreateRefreshToken() string {
	return uuid.New().String()
}

func (m *Manager) GetIDFromToken(token string) (users.ID, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.Secret), nil
	})
	if err != nil {
		var jwtErr *jwt.ValidationError
		if errors.As(err, &jwtErr); jwtErr.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return 0, ErrExpiredToken
		}
		return 0, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*jwt.StandardClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrInvalidSubject
	}

	return users.ID(id), nil
}
