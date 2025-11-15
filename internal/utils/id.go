package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// GenerateID generates a unique 8-character alphanumeric ID
func GenerateID() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// Base64 encode and take first 8 characters, make alphanumeric
	encoded := base64.URLEncoding.EncodeToString(bytes)
	// Remove non-alphanumeric characters and take first 8
	id := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, encoded)
	
	if len(id) < 8 {
		// Fallback: pad with random chars
		for len(id) < 8 {
			extra := make([]byte, 1)
			rand.Read(extra)
			char := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"[extra[0]%62]
			id += string(char)
		}
	}
	
	return id[:8], nil
}

