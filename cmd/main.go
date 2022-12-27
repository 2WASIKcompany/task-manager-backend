package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"log"
	"os"
	"task-manager-backend/internal/app/api"
	"task-manager-backend/internal/app/config"
	"task-manager-backend/internal/app/repository"
	"task-manager-backend/internal/app/repository/redis_repository"
)

type combineAuthRepository struct {
	*repository.PostgresRepository
	*redis_repository.RedisRepository
}

func authStorage(storage *repository.PostgresRepository) auth.Storage {
	rs, err := redis_repository.NewRedisRepo()
	if err != nil {
		log.Printf("redis connect err: %v", err)
		os.Exit(1)
	}
	return combineAuthStorage{
		storage, rs,
	}
}

// main godoc
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	fx.New(
		fx.Provide(
			context.Background,
			config.NewConfig,
			api.NewApi,
			gin.Default,
			repository.NewPostgresRepository,
			authStorage,
		),
		fx.Invoke(api.StartHook),
	).Run()
}
