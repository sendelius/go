package captcha

import (
	"strings"
	"time"

	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/redis"
)

const (
	ChallengeTTL    = 15 * time.Minute
	ChallengePrefix = "captcha:challenge:"
	TokenPrefix     = "captcha:token:"
)

type Service struct {
	redis *redis.Service
}

func newService(redis *redis.Service) *Service {
	return &Service{
		redis: redis,
	}
}

func (s *Service) Create() (map[string]any, error) {
	id, err := randomString(16)
	if err != nil {
		return nil, apperrors.New(ErrorChallengeCreate, "не удалось создать captcha")
	}

	challenge, err := randomString(32)
	if err != nil {
		return nil, apperrors.New(ErrorChallengeCreate, "не удалось создать captcha")
	}

	if err := s.redis.Set(ChallengePrefix+id, challenge, ChallengeTTL); err != nil {
		return nil, apperrors.New(ErrorChallengeCreate, "не удалось создать captcha")
	}

	return map[string]any{
		"id":        id,
		"challenge": challenge,
	}, nil
}

func (s *Service) Verify(id string, nonce uint64) (string, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return "", apperrors.New(ErrorVerifyNotFound, "captcha challenge не найден")
	}

	challenge, err := s.redis.GetAndDelete(ChallengePrefix + id)
	if err != nil {
		return "", apperrors.New(ErrorVerifyNotFound, "captcha challenge не найден")
	}

	hash := calculateHash(challenge, nonce)

	if !strings.HasPrefix(hash, strings.Repeat("0", 4)) {
		return "", apperrors.New(ErrorVerifyInvalid, "неверное решение captcha")
	}

	token, err := randomString(32)
	if err != nil {
		return "", apperrors.New(ErrorVerifyCheck, "не удалось проверить captcha")
	}

	if err := s.redis.Set(TokenPrefix+token, "1", ChallengeTTL); err != nil {
		return "", apperrors.New(ErrorVerifyCheck, "не удалось проверить captcha")
	}

	return token, nil
}

func (s *Service) Consume(token string) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return apperrors.New(ErrorConsumeCheck, "проверка captcha не пройдена, обновите страницу и попробуйте снова")
	}

	_, err := s.redis.GetAndDelete(TokenPrefix + token)
	if err != nil {
		return apperrors.New(ErrorConsumeCheck, "проверка captcha не пройдена, обновите страницу и попробуйте снова")
	}

	return nil
}

func (s *Service) Delete(token string) {
	token = strings.TrimSpace(token)
	if token != "" {
		_ = s.redis.Delete(TokenPrefix + token)
	}
}
