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

func FormatFingerprint(fp string) string {
	var result []string
	for i := 0; i < len(fp); i += 2 {
		if i+2 <= len(fp) {
			result = append(result, fp[i:i+2])
		}
	}
	return strings.Join(result, ":")
}

func EncodeMessage(msg Message, config *config.Config) ([]byte, error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %v", err)
	}
	if len(jsonData) > math.MaxUint32 {
		return nil, fmt.Errorf("message too large: %d bytes", len(jsonData))
	}
	if len(jsonData) > config.MaxPacketSize {
		return nil, fmt.Errorf("message too large: %d bytes", len(jsonData))
	}
	buf := make([]byte, 4+len(jsonData))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(jsonData)))
	copy(buf[4:], jsonData)
	return buf, nil
}
