package redis

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New() *Redis {
	r := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})
	return &Redis{
		Client: r,
	}
}

func (r *Redis) Set(key string, value string, expiration time.Duration) error {
	return r.Client.Set(context.Background(), key, value, expiration).Err()
}

func (r *Redis) Get(key string) (string, error) {
	return r.Client.Get(context.Background(), key).Result()
}

func (r *Redis) GetAndDelete(key string) (string, error) {
	return r.Client.GetDel(context.Background(), key).Result()
}

func (r *Redis) Delete(key string) error {
	return r.Client.Del(context.Background(), key).Err()
}

func (r *Redis) Exists(key string) (bool, error) {
	result, err := r.Client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *Redis) Expire(key string, expiration time.Duration) error {
	return r.Client.Expire(context.Background(), key, expiration).Err()
}

func (r *Redis) TTL(key string) (time.Duration, error) {
	return r.Client.TTL(context.Background(), key).Result()
}

func (r *Redis) Ping() error {
	return r.Client.Ping(context.Background()).Err()
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
