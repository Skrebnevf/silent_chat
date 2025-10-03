package utils

import (
	"crypto/rand"
	"time"

	"github.com/briandowns/spinner"
)

// Spinner displays an animated spinner with the given message for the specified duration.
// Uses the briandowns/spinner library for a more professional and feature-rich spinner.
func Spinner(message string, duration time.Duration) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Use dots spinner for better visibility
	s.Suffix = " " + message
	s.Start()

	time.Sleep(duration)

	s.Stop()
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
