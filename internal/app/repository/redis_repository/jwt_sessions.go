package redis_repository

import (
	"strconv"
	"time"
)

const (
	RefreshTemplateKey = "refresh_"
)

func (r *RedisRepository) CashRefreshToken(uid uint64, refresh string, at time.Duration) error {
	return r.db.Set(RefreshTemplateKey+strconv.Itoa(int(uid)), refresh, at).Err()
}

func (r *RedisRepository) GetRefreshToken(uid uint64) (string, error) {
	return r.db.Get(RefreshTemplateKey + strconv.Itoa(int(uid))).Result()
}
