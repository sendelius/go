package captcha

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sendelius/go/redis"
)

const (
	ChallengeTTL    = 15 * time.Minute
	ChallengePrefix = "captcha:challenge:"
	TokenPrefix     = "captcha:token:"
)

type Captcha struct {
	redis *redis.Redis
}

func New(redis *redis.Redis) *Captcha {
	return &Captcha{
		redis: redis,
	}
}

func (c *Captcha) Create() (map[string]any, error) {
	id, err := randomString(16)
	if err != nil {
		return nil, err
	}

	challenge, err := randomString(32)
	if err != nil {
		return nil, err
	}

	if err := c.redis.Set(ChallengePrefix+id, challenge, ChallengeTTL); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":        id,
		"challenge": challenge,
	}, nil
}

func (c *Captcha) Verify(id string, nonce uint64) (string, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return "", errors.New("captcha challenge не найден")
	}

	challenge, err := c.redis.GetAndDelete(ChallengePrefix + id)
	if err != nil {
		return "", errors.New("captcha challenge не найден")
	}

	hash := calculateHash(challenge, nonce)

	if !strings.HasPrefix(hash, strings.Repeat("0", 4)) {
		return "", errors.New("неверное решение captcha")
	}

	token, err := randomString(32)
	if err != nil {
		return "", err
	}

	if err := c.redis.Set(TokenPrefix+token, "1", ChallengeTTL); err != nil {
		return "", err
	}

	return token, nil
}

func (c *Captcha) Consume(token string) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return errors.New("проверка captcha не пройдена, обновите страницу и попробуйте снова")
	}

	_, err := c.redis.Get(TokenPrefix + token)
	if err != nil {
		return errors.New("проверка captcha не пройдена, обновите страницу и попробуйте снова")
	}

	return nil
}

func (c *Captcha) Delete(token string) {
	token = strings.TrimSpace(token)
	if token != "" {
		_ = c.redis.Delete(TokenPrefix + token)
	}
}

func calculateHash(challenge string, nonce uint64) string {
	input := challenge + ":" + strconv.FormatUint(nonce, 10)
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func randomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("ошибка генерации случайного числа: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
