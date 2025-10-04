package utils

import (
	"strings"
	"testing"
	"time"
)

func TestSpinner(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		duration time.Duration
	}{
		{
			name:     "short spinner",
			message:  "Loading...",
			duration: 100 * time.Millisecond,
		},
		{
			name:     "empty message",
			message:  "",
			duration: 50 * time.Millisecond,
		},
		{
			name:     "long message",
			message:  "This is a very long message for testing",
			duration: 200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			Spinner(tt.message, tt.duration)
			elapsed := time.Since(start)
			if elapsed < tt.duration ||
				elapsed > tt.duration+50*time.Millisecond {
				t.Errorf(
					"Spinner duration mismatch: expected ~%v, got %v",
					tt.duration,
					elapsed,
				)
			}
		})
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid length 10",
			length:  10,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			wantErr: false,
		},
		{
			name:    "large length",
			length:  100,
			wantErr: false,
		},
		{
			name:    "negative length",
			length:  -1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := RandomString(tt.length)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				if tt.errMsg != "" &&
					!strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf(
						"Expected error containing %q, got %q",
						tt.errMsg,
						err.Error(),
					)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(str) != tt.length {
					t.Errorf("Expected length %d, got %d", tt.length, len(str))
				}
				charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
				for _, char := range str {
					if !strings.ContainsRune(charset, char) {
						t.Errorf("Invalid character in random string: %c", char)
					}
				}
			}
		})
	}
}
