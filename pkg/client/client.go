package client

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"silent_chat/internal/utils"
	"silent_chat/pkg/config"
	"silent_chat/pkg/protocol"

	"golang.org/x/term"
)

type Client struct {
	Conn        *tls.Conn
	Reader      *bufio.Reader
	Addr        string
	Username    string
	Config      *config.Config
	Mutex       sync.Mutex
	OutputMutex sync.Mutex
	Connected   bool
}

func (c *Client) ReadMessage() (protocol.Message, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(c.Conn, header); err != nil {
		return protocol.Message{}, fmt.Errorf("failed to read header: %v", err)
	}
	size := binary.BigEndian.Uint32(header)

	if c.Config.MaxPacketSize > math.MaxUint32 {
		return protocol.Message{}, fmt.Errorf(
			"max packet size exceeds maximum allowed value",
		)
	}
	if size > uint32(c.Config.MaxPacketSize) {
		return protocol.Message{}, fmt.Errorf("packet too large: %d", size)
	}

	body := make([]byte, size)
	if _, err := io.ReadFull(c.Conn, body); err != nil {
		return protocol.Message{}, fmt.Errorf("failed to read body: %v", err)
	}

	var msg protocol.Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return protocol.Message{}, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return msg, nil
}

func (c *Client) WriteMessage(msg protocol.Message) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if !c.Connected {
		return fmt.Errorf("not connected")
	}

	data, err := protocol.EncodeMessage(msg, c.Config)
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(data)
	if err != nil {
		c.Connected = false
		return fmt.Errorf("failed to write message: %v", err)
	}
	return nil
}

func (c *Client) SendFakeMessage() error {
	randInt := rand.Intn(19) + 1
	randString, err := utils.RandomString(randInt)
	if err != nil {
		return fmt.Errorf("failed to generate random string: %v", err)
	}
	fake := protocol.Message{
		Type: "fake",
		Text: randString,
	}

	return c.WriteMessage(fake)
}

func (c *Client) Connect() error {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", c.Addr, config)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}

	if err := protocol.VerifyFingerprint(conn, c.Config.ExpectedFP); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("failed to close connection: %v", closeErr)
		}
		return err
	}

	fmt.Printf("Connected to %s\n", c.Addr)

	fmt.Print("Enter server password: ")
	password, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Printf("cannot read password: %v", err)
		os.Exit(1)
	}
	passwordString := string(password)

	fmt.Print("\nEnter your username: ")
	username, _ := c.Reader.ReadString('\n')
	username = strings.TrimSpace(username)

	c.Username = username

	c.Conn = conn
	c.Connected = true

	authMsg := protocol.Message{
		Type:     "auth",
		Password: passwordString,
		Username: username,
	}
	if err := c.WriteMessage(authMsg); err != nil {
		c.Connected = false
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("failed to close connection: %v", closeErr)
		}

		return err
	}

	if setDeadlineErr := conn.SetReadDeadline(time.Now().Add(c.Config.AuthTimeout)); setDeadlineErr != nil {
		log.Printf("read deadline err: %v", setDeadlineErr)
	}
	resp, err := c.ReadMessage()
	if setDeadlineErr := conn.SetReadDeadline(time.Time{}); setDeadlineErr != nil {
		log.Printf("read deadline err: %v", setDeadlineErr)
	}
	if err != nil {
		c.Connected = false
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("failed to close connection: %v", closeErr)
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("authentication timeout")
		}
		return fmt.Errorf("authentication error: %v", err)
	}

	if resp.Type == "auth_result" {
		if resp.Success {
			fmt.Printf("Authentication successful. Username: %s\n", username)
			fmt.Println(
				"You can now send messages. Type '/quit' or '/q' to exit.",
			)
			return nil
		}
		fmt.Printf("Authentication failed: %s\n", resp.Error)

		c.Connected = false
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("failed to close connection: %v", closeErr)
		}

		return fmt.Errorf("authentication failed: %s", resp.Error)
	}

	c.Connected = false
	if closeErr := conn.Close(); closeErr != nil {
		log.Printf("failed to close connection: %v", closeErr)
	}

	return fmt.Errorf("unexpected response from server")
}

func (c *Client) Listen() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Listen goroutine recovered from panic: %v\n", r)
		}
	}()

	for {
		if !c.Connected {
			return
		}

		if setDeadlineErr := c.Conn.SetReadDeadline(time.Now().Add(c.Config.ReadTimeout)); setDeadlineErr != nil {
			log.Printf("set deadline err: %v", setDeadlineErr)
		}
		msg, err := c.ReadMessage()
		if setDeadlineErr := c.Conn.SetReadDeadline(time.Time{}); setDeadlineErr != nil {
			log.Printf("set deadline err: %v", setDeadlineErr)
		}

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if strings.Contains(err.Error(), "EOF") ||
				strings.Contains(
					err.Error(),
					"use of closed network connection",
				) {
				fmt.Println("\nConnection closed by server")
				c.Connected = false
				return
			}
			fmt.Printf("\nListen error: %v\n", err)
			c.Connected = false
			return
		}

		if msg.Type == "chat" && msg.Text != "" && msg.SenderName != "" {
			if msg.SenderIP != "" {
				fmt.Printf("\r%s: %s\n> ", msg.SenderName, msg.Text)
			} else {
				fmt.Printf("\r%s: %s\n> ", msg.SenderName, msg.Text)
			}
		}
	}
}

func (c *Client) SendLoop() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Send goroutine recovered from panic: %v\n", r)
		}
	}()

	for {

		if !c.Connected {
			return fmt.Errorf("connection lost")
		}

		c.OutputMutex.Lock()
		fmt.Print("> ")
		c.OutputMutex.Unlock()
		msg, err := c.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Disconnecting...")
				return nil
			}
			return err
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		if msg == "/quit" || msg == "/q" {
			fmt.Println("The Lord loves the patient, and internet privacy")
			os.Exit(0)
			return nil
		}

		if msg == "/kill" {
			fmt.Println("Sending server shutdown command...")
			// Отправляем специальные байты: "KILL" + 4 нулевых байта
			killBytes := []byte("MSGE\x00\x00\x00\x00")
			if _, err := c.Conn.Write(killBytes); err != nil {
				fmt.Printf("Failed to send kill command: %v\n", err)
			}
			return nil
		}

		msgData := protocol.Message{
			Type:       "chat",
			Text:       msg,
			SenderName: c.Username,
		}

		delay := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(delay)

		if err := c.WriteMessage(msgData); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
			c.Connected = false
			return err
		}
	}
}

func (c *Client) Run() error {
	retryCount := 0

	for {
		err := c.Connect()
		if err != nil {
			retryCount++

			if strings.Contains(err.Error(), "authentication failed") {

				fmt.Printf(
					"Authentication failed. Retrying in %v...\n",
					c.Config.AuthFailDelay,
				)
				utils.Spinner("Waiting", c.Config.AuthFailDelay)
				continue
			}

			if retryCount >= c.Config.MaxRetries {
				backoffDelay := c.Config.ReconnectDelay + c.Config.BackoffIncrement*time.Duration(
					retryCount-c.Config.MaxRetries,
				)

				fmt.Printf(
					"Max retries reached. Backing off for %v...\n",
					backoffDelay,
				)
				utils.Spinner("Reconnecting", backoffDelay)
			} else {

				fmt.Printf("Connection failed. Retrying in %v... (attempt %d/%d)\n",
					c.Config.ReconnectDelay, retryCount, c.Config.MaxRetries)
				utils.Spinner("Reconnecting", c.Config.ReconnectDelay)
			}
			continue
		}

		retryCount = 0

		delay := time.Duration(rand.Intn(30)) * time.Second

		go func() {
			ticker := time.NewTicker(delay)
			defer ticker.Stop()
			for {
				<-ticker.C
				if c.Connected {
					if err := c.SendFakeMessage(); err != nil {
						log.Printf("send fake message, err: %v", err)
					}
				}
			}
		}()

		var wg sync.WaitGroup

		wg.Go(func() {
			c.Listen()
		})
		err = c.SendLoop()

		wg.Wait()

		if c.Conn != nil {
			if closeErr := c.Conn.Close(); closeErr != nil {
				log.Printf("failed connecting close: %v", closeErr)
			}
			c.Connected = false
		}

		if err == nil {
			return nil // Clean exit
		}

		fmt.Printf("Session ended: %v\n", err)
		fmt.Printf("Reconnecting in %v...\n", c.Config.ReconnectDelay)
		utils.Spinner("Reconnecting", c.Config.ReconnectDelay)
	}
}
