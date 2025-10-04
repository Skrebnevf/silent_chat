package config

import "time"

// Config represents the configuration settings for the chat application.
// It includes network parameters, timeouts, retry strategies, and security settings.
type Config struct {
	MaxPacketSize         uint32        // Maximum packet size in bytes (default 65536)
	AbsoluteMaxPacketSize uint32        // Absolute maximum packet size (default 10MB)
	AuthTimeout           time.Duration // Authentication timeout (default 5 seconds)
	ReconnectDelay        time.Duration // Delay before reconnecting (default 5 seconds)
	AuthFailDelay         time.Duration // Delay after authentication failure (default 1 second)
	ReadTimeout           time.Duration // Read timeout (default 0 - no timeout)
	MaxRetries            int           // Maximum number of retries (default 5)
	BackoffIncrement      time.Duration // Backoff increment (default 2 seconds)
	ExpectedFP            string        // Expected certificate fingerprint for verification
	Addr                  string        // Server address for connection
	DialTimeout           time.Duration // Timeout for TLS dial (default 15 seconds)
}

// NewConfig creates a new Config instance with default values.
// It initializes all fields with sensible defaults for a chat application.
func NewConfig() *Config {
	return &Config{
		MaxPacketSize:         65536,
		AbsoluteMaxPacketSize: 10 * 1024 * 1024,
		AuthTimeout:           5 * time.Second,
		ReconnectDelay:        5 * time.Second,
		AuthFailDelay:         1 * time.Second,
		ReadTimeout:           0,
		MaxRetries:            5,
		BackoffIncrement:      2 * time.Second,
		DialTimeout:           15 * time.Second,
	}
}
