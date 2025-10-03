package protocol

import (
	"strings"
	"testing"

	"silent_chat/pkg/config"
)

func TestFormatFingerprint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "even length hex string",
			input:    "aabbccdd",
			expected: "aa:bb:cc:dd",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single byte",
			input:    "aa",
			expected: "aa",
		},
		{
			name:     "odd length hex string",
			input:    "aaa",
			expected: "aa",
		},
		{
			name:     "long hex string",
			input:    "aabbccddeeff",
			expected: "aa:bb:cc:dd:ee:ff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFingerprint(tt.input)
			if result != tt.expected {
				t.Errorf(
					"FormatFingerprint(%q) = %q, want %q",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestEncodeMessage(t *testing.T) {
	cfg := &config.Config{
		MaxPacketSize: 1024,
	}

	tests := []struct {
		name      string
		message   Message
		config    *config.Config
		wantErr   bool
		errString string
	}{
		{
			name: "normal message",
			message: Message{
				Type: "chat",
				Text: "Hello",
			},
			config:  cfg,
			wantErr: false,
		},
		{
			name: "message too large",
			message: Message{
				Type: "chat",
				Text: strings.Repeat("a", int(cfg.MaxPacketSize)+1),
			},
			config:    cfg,
			wantErr:   true,
			errString: "message too large",
		},
		{
			name: "message at max size",
			message: Message{
				Type: "chat",
				Text: strings.Repeat("a", 900),
			},
			config:  cfg,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := EncodeMessage(tt.message, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"EncodeMessage() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errString) {
				t.Errorf(
					"EncodeMessage() error = %v, want error containing %q",
					err,
					tt.errString,
				)
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("EncodeMessage() returned empty data for valid message")
			}
		})
	}
}
