package api

import (
	"math/rand"
)

// charset contains all Base62 characters: 0-9, a-z, A-Z
const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// codeLength is the length of generated short codes.
const codeLength = 6

// GenerateShortCode creates a random 6-character Base62 code.
// Example output: "aZ91Xk"
func GenerateShortCode() string {
	code := make([]byte, codeLength)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}
