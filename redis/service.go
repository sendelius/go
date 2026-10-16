package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/env"
)

type Service struct {
	client *goredis.Client
	prefix string
}

func newService() *Service {
	return &Service{
		client: goredis.NewClient(&goredis.Options{
			Addr: env.String("REDIS_HOST") + ":" + env.String("REDIS_PORT"),
		}),
		prefix: env.String("REDIS_PREFIX"),
	}
}

func (s *Service) key(key string) string {
	return s.prefix + key
}

func (s *Service) Set(key string, value string, expiration time.Duration) error {
	if err := s.client.Set(context.Background(), s.key(key), value, expiration).Err(); err != nil {
		return apperrors.New(ErrorSet, "ошибка записи в Redis")
	}
	return nil
}

func (s *Service) Get(key string) (string, error) {
	value, err := s.client.Get(context.Background(), s.key(key)).Result()
	if err != nil {
		return "", apperrors.New(ErrorGet, "ошибка чтения из Redis")
	}
	return value, nil
}

func (s *Service) GetAndDelete(key string) (string, error) {
	value, err := s.client.GetDel(context.Background(), s.key(key)).Result()
	if err != nil {
		return "", apperrors.New(ErrorGetAndDelete, "ошибка получения и удаления ключа в Redis")
	}
	return value, nil
}

func (s *Service) Delete(key string) error {
	if err := s.client.Del(context.Background(), s.key(key)).Err(); err != nil {
		return apperrors.New(ErrorDelete, "ошибка удаления ключа в Redis")
	}
	return nil
}

func (s *Service) Exists(key string) (bool, error) {
	result, err := s.client.Exists(context.Background(), s.key(key)).Result()
	if err != nil {
		return false, apperrors.New(ErrorExists, "ошибка проверки существования ключа в Redis")
	}
	return result > 0, nil
}

func (s *Service) Expire(key string, expiration time.Duration) error {
	if err := s.client.Expire(context.Background(), s.key(key), expiration).Err(); err != nil {
		return apperrors.New(ErrorExpire, "ошибка установки срока действия ключа в Redis")
	}
	return nil
}

func (s *Service) TTL(key string) (time.Duration, error) {
	result, err := s.client.TTL(context.Background(), s.key(key)).Result()
	if err != nil {
		return 0, apperrors.New(ErrorTTL, "ошибка получения TTL ключа в Redis")
	}
	return result, nil
}

func (s *Service) Ping() error {
	if err := s.client.Ping(context.Background()).Err(); err != nil {
		return apperrors.New(ErrorPing, "ошибка проверки соединения с Redis")
	}
	return nil
}

func (s *Service) Close() error {
	if err := s.client.Close(); err != nil {
		return apperrors.New(ErrorClose, "ошибка закрытия соединения с Redis")
	}
	return nil
}
