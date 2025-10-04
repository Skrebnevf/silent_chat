// Package protocol provides structures and functions for handling chat messages,
// including TLS certificate verification, fingerprint formatting, and message encoding.
package protocol

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"silent_chat/pkg/config"
)

// Message represents a protocol message used in the chat application for communication between client and server.
// It includes fields for message type, content, sender information, authentication, and status.
type Message struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	SenderName string `json:"sender_name,omitempty"`
	SenderIP   string `json:"sender_ip,omitempty"`
	Password   string `json:"password,omitempty"`
	Username   string `json:"username,omitempty"`
	Success    bool   `json:"success,omitempty"`
	Error      string `json:"error,omitempty"`
}

// VerifyFingerprint verifies the TLS certificate fingerprint against an expected value to ensure secure connection.
// It performs a handshake, extracts the server's certificate, computes its SHA256 fingerprint, and compares it with the expected one.
// If expectedFP is empty, it warns and allows insecure connection for development purposes.
func VerifyFingerprint(conn *tls.Conn, expectedFP string) error {
	if err := conn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %v", err)
	}
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return fmt.Errorf("no server certificates found")
	}
	actualFP := sha256.Sum256(certs[0].Raw)
	actualFPHex := strings.ToLower(hex.EncodeToString(actualFP[:]))
	if expectedFP == "" {
		fmt.Printf("\nWARNING: Connecting without certificate verification!\n")
		fmt.Printf("Server fingerprint: %s\n", FormatFingerprint(actualFPHex))
		fmt.Printf(
			"For secure connections, set:\nexport CHAT_SERVER_FINGERPRINT=%s\n\n",
			actualFPHex,
		)
		return nil
	}

	if actualFPHex == expectedFP {
		fmt.Println("Certificate verified successfully!")
		return nil
	}

	fmt.Printf("Certificate fingerprint mismatch!\n")
	fmt.Printf("Expected: %s\n", expectedFP)
	fmt.Printf("Actual:   %s\n", actualFPHex)
	return fmt.Errorf("fingerprint mismatch")
}

// FormatFingerprint formats a hexadecimal fingerprint string into a colon-separated format for better readability.
// It groups the hex string into pairs separated by colons (e.g., "aabbccdd" becomes "aa:bb:cc:dd").
func FormatFingerprint(fp string) string {
	var result []string
	for i := 0; i < len(fp); i += 2 {
		if i+2 <= len(fp) {
			result = append(result, fp[i:i+2])
		}
	}
	return strings.Join(result, ":")
}

// EncodeMessage encodes a Message struct into a JSON byte array with a 4-byte length prefix for network transmission.
// It marshals the message to JSON, checks size limits from config, and prepends the length as a big-endian uint32.
// Returns an error if the message exceeds MaxPacketSize or other encoding issues occur.
func EncodeMessage(msg Message, config *config.Config) ([]byte, error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %v", err)
	}
	if len(jsonData) > math.MaxUint32 {
		return nil, fmt.Errorf("message too large: %d bytes", len(jsonData))
	}
	if uint32(len(jsonData)) > config.MaxPacketSize {
		return nil, fmt.Errorf("message too large: %d bytes", len(jsonData))
	}
	buf := make([]byte, 4+len(jsonData))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(jsonData)))
	copy(buf[4:], jsonData)
	return buf, nil
}
