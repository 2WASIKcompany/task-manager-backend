package redis_repository

import "github.com/go-redis/redis"

func NewRedisRepo(addr string, passwd string, dbID int) (*RedisRepository, error) {
	testClient := &RedisRepository{
		db: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: passwd,
			DB:       dbID,
		}),
	}

	_, err := testClient.db.Ping().Result()

	return testClient, err

}

type RedisRepository struct {
	db *redis.Client
}
