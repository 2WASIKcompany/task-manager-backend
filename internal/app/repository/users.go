package repository

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"task-manager-backend/internal/app/models/users"
)

const (
	usersTable = "users"
	ID         = "id"
	Email      = "email"
	PwdHash    = "pwd_hash"
	Status     = "status"
	Confirmed  = "confirmed"
	RegData    = "reg_date"
)

func (p *PostgresRepository) CreateUser(ctx context.Context, user users.User) error {
	query, args, err := sq.
		Insert(usersTable).
		Columns(ID, Email, PwdHash, Status).
		Values(
			user.ID,
			user.Email,
			user.PwdHash,
			user.Status,
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, query, args...)
	return err
}

func (p *PostgresRepository) GetUserByEmail(ctx context.Context, email users.Email) (users.User, error) {
	var user users.User

	query, args, err := sq.
		Select(ID, Email, PwdHash, Status, Confirmed).
		From(usersTable).
		Where(
			sq.Eq{
				Email: email,
			},
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return user, err
	}

	err = p.db.GetContext(ctx, &user, query, args...)
	if err == sql.ErrNoRows {
		err = nil
	}
	return user, err
}

func (p *PostgresRepository) GetUserByUserID(ctx context.Context, userID users.ID) (users.User, error) {
	var user users.User

	query, args, err := sq.
		Select(ID, Email, PwdHash, Status, Confirmed).
		From(usersTable).
		Where(
			sq.Eq{
				ID: userID,
			},
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return user, err
	}

	err = p.db.GetContext(ctx, &user, query, args...)
	if err == sql.ErrNoRows {
		err = nil
	}
	return user, err
}

func (p *PostgresRepository) ConfirmUser(ctx context.Context, userID users.ID) error {
	query, args, err := sq.
		Update(usersTable).
		Set(Confirmed, true).
		Where(
			sq.Eq{
				ID: userID,
			},
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, query, args...)
	return err

}

func (p *PostgresRepository) ChangePasswordByUserID(ctx context.Context, userID users.ID, newPassword string) error {
	query, args, err := sq.
		Update(usersTable).
		Set(PwdHash, newPassword).
		Where(
			sq.Eq{
				ID: userID,
			},
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, query, args...)
	return err
}
