package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"task-manager-backend/internal/app/config"
)

func NewPostgresRepository(cfg config.ServiceConfiguration) *PostgresRepository {
	db, err := sqlx.Connect("postgres", cfg.PostgresDSN.String())
	if err != nil {
		log.Fatalln(err)
	}

	return &PostgresRepository{
		db: db,
	}
}

type PostgresRepository struct {
	db *sqlx.DB
}
