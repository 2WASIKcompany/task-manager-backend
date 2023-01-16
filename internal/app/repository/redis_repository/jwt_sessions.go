package redis_repository

import (
	"strconv"
	"task-manager-backend/internal/app/models/users"
	"time"
)

const (
	RefreshTemplateKey = "refresh_"
)

func (r *RedisRepository) CashRefreshToken(uid users.ID, refresh string, at time.Duration) error {
	return r.db.Set(RefreshTemplateKey+refresh, strconv.Itoa(int(uid)), at).Err()
}

func (r *RedisRepository) GetUserIDByRefreshToken(refresh string) (string, error) {
	return r.db.Get(RefreshTemplateKey + refresh).Result()
}

func (r *RedisRepository) DeleteSession(refresh string) error {
	return r.db.Del(RefreshTemplateKey + refresh).Err()
}
