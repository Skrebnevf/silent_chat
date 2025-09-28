package utils

import (
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
