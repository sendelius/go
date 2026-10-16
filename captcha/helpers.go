package captcha

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

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
