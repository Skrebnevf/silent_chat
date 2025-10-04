package utils

import (
	"crypto/rand"
	"time"

	"github.com/briandowns/spinner"
)

// Spinner displays an animated spinner with the given message for the specified duration.
// It blocks the current goroutine until the duration elapses, making it suitable for simple CLI feedback.
// Uses the briandowns/spinner library for a more professional and feature-rich spinner.
func Spinner(message string, duration time.Duration) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	time.Sleep(duration)

	s.Stop()
}

// RandomString generates a random string of the specified length using cryptographically secure random bytes.
// The string consists of alphanumeric characters (a-z, A-Z, 0-9).
// Returns an error if random byte generation fails.
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
