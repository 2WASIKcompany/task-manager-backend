package redis_repository

import "github.com/go-redis/redis"

func NewRedisRepo(addr string, passwd string, dbID int) *RedisRepository {
	return &RedisRepository{
		db: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: passwd,
			DB:       dbID,
		}),
	}
}

type RedisRepository struct {
	db *redis.Client
}
