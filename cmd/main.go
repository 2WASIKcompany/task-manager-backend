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
	"task-manager-backend/internal/app/service/auth"
	"task-manager-backend/internal/app/service/mail"
)

type combineAuthRepository struct {
	*repository.PostgresRepository
	*redis_repository.RedisRepository
}

func authStorage(storage *repository.PostgresRepository) auth.Repository {
	rs, err := redis_repository.NewRedisRepo(config.Load().RedisAddr, config.Load().RedisPasswd, 0)
	if err != nil {
		log.Printf("redis connect err: %v", err)
		os.Exit(1)
	}
	return combineAuthRepository{
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
			mail.NewSender,
			authStorage,
			auth.NewManger,
			auth.NewService,
		),
		fx.Invoke(api.StartHook),
	).Run()
}
