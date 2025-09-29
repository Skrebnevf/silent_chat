package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

func Spinner(message string, duration time.Duration) {
	spinnerChars := []string{"|", "/", "-", "\\"}
	start := time.Now()
	for time.Since(start) < duration {
		for _, char := range spinnerChars {
			fmt.Printf("\r%s %s", message, char)
			time.Sleep(100 * time.Millisecond)
			if time.Since(start) >= duration {
				break
			}
		}
	}
	fmt.Printf("\r%s Done\n", message)
}

func RandomString(length int) (string, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}
