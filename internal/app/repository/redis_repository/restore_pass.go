package redis_repository

import (
	"errors"
	"strconv"
	"task-manager-backend/internal/app/models/users"
	"time"
)

const (
	restoreEmailRate = "restore_rate:"
	restoreUID       = "restore_uid:"
)

func (r *RedisRepository) CreateRestoreUID(email users.Email, userID users.ID, uid string) error {
	key := restoreEmailRate + string(email)
	result, err := r.db.Exists(key).Result()
	if err != nil {
		return err
	}
	if result == 0 {
		if r.db.Set(key, 1, 86400*time.Second).Err() != nil {
			return errors.New("CreateRestoreUID err: create restore_rate err")
		}
	} else if rate, err := r.db.Get(key).Result(); err == nil && rate != "2" {
		if r.db.Incr(key).Err() != nil {
			return errors.New("CreateRestoreUID err: incr err")
		}
	} else {
		return errors.New("CreateRestoreUID err:request limit exceeded")
	}

	if r.db.Set(restoreUID+uid, int64(userID), 10800*time.Second).Err() != nil {
		return errors.New("CreateRestoreUID err: create restore uid err")
	}

	return nil
}

func (r *RedisRepository) GetUserIDByRestoreUID(uid string) (users.ID, error) {
	request, err := r.db.Get(restoreUID + uid).Result()
	if err != nil {
		return 0, errors.New("GetUserIDByRestoreUID err: restore_uid not found")
	}

	userID, err := strconv.Atoi(request)
	if err != nil {
		return 0, errors.New("GetUserIDByRestoreUID err: bad user_id")
	}

	r.db.Del(restoreUID + uid)

	return users.ID(userID), nil
}
