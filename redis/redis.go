package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sendelius/go/env"
)

type Redis struct {
	Client *redis.Client
	prefix string
}

func New() *Redis {
	r := redis.NewClient(&redis.Options{
		Addr: env.String("REDIS_HOST") + ":" + env.String("REDIS_PORT"),
	})
	return &Redis{
		Client: r,
		prefix: env.String("REDIS_PREFIX"),
	}
}

func (r *Redis) key(key string) string {
	return r.prefix + key
}

func (r *Redis) Set(key string, value string, expiration time.Duration) error {
	return r.Client.Set(context.Background(), r.key(key), value, expiration).Err()
}

func (r *Redis) Get(key string) (string, error) {
	return r.Client.Get(context.Background(), r.key(key)).Result()
}

func (r *Redis) GetAndDelete(key string) (string, error) {
	return r.Client.GetDel(context.Background(), r.key(key)).Result()
}

func (r *Redis) Delete(key string) error {
	return r.Client.Del(context.Background(), r.key(key)).Err()
}

func (r *Redis) Exists(key string) (bool, error) {
	result, err := r.Client.Exists(context.Background(), r.key(key)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *Redis) Expire(key string, expiration time.Duration) error {
	return r.Client.Expire(context.Background(), r.key(key), expiration).Err()
}

func (r *Redis) TTL(key string) (time.Duration, error) {
	return r.Client.TTL(context.Background(), r.key(key)).Result()
}

func (r *Redis) Ping() error {
	return r.Client.Ping(context.Background()).Err()
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
