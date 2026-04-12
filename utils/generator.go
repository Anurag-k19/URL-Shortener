package utils

import "math/rand/v2"

// GenerateShortId creates a random 6-character alphanumeric string for short URLs
func GenerateShortId() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 6)

	for i := range id {
		id[i] = charset[rand.UintN(uint(len(charset)))]
	}

	return string(id)
}
