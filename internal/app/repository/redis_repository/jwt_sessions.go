package redis_repository

import (
	"time"
)

const (
	restoreEmailRate = "restore_rate:"
	restoreUID       = "restore_uid:"
)

func (r *RedisRepository) CreateJWTSession(token, refresh string, at time.Time) error {

}
