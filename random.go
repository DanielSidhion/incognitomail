package incognitomail

import (
	"crypto/rand"
	"math"
)

const (
	allowedCharacters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	allowedCharactersNum = len(allowedCharacters)
)

var (
	// How many bits we actually need to use to represent an index from allowedCharacters
	bitsPerIndex = uint(math.Ceil(math.Log2(float64(allowedCharactersNum))))

	indexMask = uint8(1<<bitsPerIndex - 1)
)

// generateRandomString will return a random string with length equals to size.
func generateRandomString(size int) (string, error) {
	buf := make([]byte, size)
	result := make([]byte, size)

	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	// TODO: better random string generation
	for i := 0; i < size; i++ {
		idx := int(uint8(buf[i])&indexMask) % allowedCharactersNum
		result[i] = allowedCharacters[idx]
	}

	return string(result), nil
}
