package config

import "time"

type Config struct {
	MaxPacketSize    int
	AuthTimeout      time.Duration
	ReconnectDelay   time.Duration
	AuthFailDelay    time.Duration
	ReadTimeout      time.Duration
	MaxRetries       int
	BackoffIncrement time.Duration
	ExpectedFP       string
	Addr             string
}

func NewConfig() *Config {
	return &Config{
		MaxPacketSize:    65536,
		AuthTimeout:      5 * time.Second,
		ReconnectDelay:   5 * time.Second,
		AuthFailDelay:    1 * time.Second,
		ReadTimeout:      30 * time.Minute,
		MaxRetries:       5,
		BackoffIncrement: 2 * time.Second,
	}
}
