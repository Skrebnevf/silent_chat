package client

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"silent_chat/internal/utils"
	"silent_chat/pkg/config"
	"silent_chat/pkg/protocol"
	"silent_chat/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const clearScreen = "\033[2J\033[H"

type Client struct {
	Conn        *tls.Conn
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
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return protocol.Message{}, fmt.Errorf(
				"header read timeout: %v",
				err,
			)
		}
		return protocol.Message{}, fmt.Errorf("failed to read header: %v", err)
	}

	size := binary.BigEndian.Uint32(header)
	if size == 0 {
		return protocol.Message{}, fmt.Errorf("packet size is zero")
	}
	if size > c.Config.MaxPacketSize {
		return protocol.Message{}, fmt.Errorf("packet too large: %d", size)
	}
	if size > c.Config.AbsoluteMaxPacketSize {
		return protocol.Message{}, fmt.Errorf(
			"packet too large or attack detected: %v",
			size,
		)
	}

	body := make([]byte, size)
	if _, err := io.ReadFull(c.Conn, body); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return protocol.Message{}, fmt.Errorf("body read timeout: %v", err)
		}
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
	n, err := c.Conn.Write(data)
	if err != nil {
		c.Connected = false
		return fmt.Errorf("failed to write message: %v", err)
	}
	if n != len(data) {
		c.Connected = false
		return fmt.Errorf("incomplete write: wrote %d bytes out of %d",
			n, len(data))
	}
	return nil
}

func (c *Client) SendFakeMessage() error {
	randInt := rand.Intn(39) + 1
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
	model := ui.NewAuthModel()
	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("auth UI error: %v", err)
	}

	authModel, ok := result.(ui.AuthModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	authData, err := authModel.GetAuthData()
	if err != nil {
		os.Exit(1)
	}

	c.Addr = authData.Host + ":" + authData.Port

	c.Username = authData.Username

	passwordString := authData.Password
	passwordBytes := []byte(passwordString)

	hash := sha256.Sum256(passwordBytes)
	hashedPassword := base64.StdEncoding.EncodeToString(hash[:])

	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	dialer := &net.Dialer{Timeout: c.Config.DialTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", c.Addr, config)
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

	c.Conn = conn
	c.Connected = true

	authMsg := protocol.Message{
		Type:     "auth",
		Password: hashedPassword,
		Username: c.Username,
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
			fmt.Printf("Authentication successful. Username: %s\n", c.Username)
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

func (c *Client) Listen(p *tea.Program) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Listen goroutine recovered from panic: %v\n", r)
		}
	}()

	for {
		if !c.Connected {
			return
		}

		var msg protocol.Message
		var err error

		if c.Config.ReadTimeout > 0 {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				c.Config.ReadTimeout,
			)
			defer cancel()

			msgChan := make(chan protocol.Message, 1)
			errChan := make(chan error, 1)

			go func() {
				msg, err := c.ReadMessage()
				if err != nil {
					errChan <- err
				} else {
					msgChan <- msg
				}
			}()

			select {
			case msg = <-msgChan:
			case err = <-errChan:
			case <-ctx.Done():
				err = ctx.Err()
			}
		} else {
			msg, err = c.ReadMessage()
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
				p.Quit()
				return
			}

			fmt.Printf("\nNetwork error: %v\n", err)
			c.Connected = false
			return
		}

		if msg.Type == "chat" && msg.Text != "" && msg.SenderName != "" {
			// Отправляем в UI
			p.Send(ui.NewChatMsg{
				Sender: msg.SenderName,
				Text:   msg.Text,
			})
		}
	}
}

func (c *Client) Run() error {
	retryCount := 0

	for {
		fmt.Print(clearScreen)
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
				utils.Spinner("Reconnecting ...", backoffDelay)
			} else {
				fmt.Printf("Connection failed. Retrying in %v... (attempt %d/%d)\n",
					c.Config.ReconnectDelay, retryCount, c.Config.MaxRetries)
				utils.Spinner("Reconnecting", c.Config.ReconnectDelay)
			}
			continue
		}

		retryCount = 0

		delay := time.Duration(rand.Intn(30)+1) * time.Second

		stopFake := make(chan struct{})

		go func() {
			ticker := time.NewTicker(delay)
			defer ticker.Stop()
			for {
				select {
				case <-stopFake:
					return
				case <-ticker.C:
					if c.Connected {
						if err := c.SendFakeMessage(); err != nil {
							log.Printf("send fake message, err: %v", err)
						}
					}
				}
			}
		}()

		chatModel := ui.NewChatModel(c.Username, func(text string) {
			msgData := protocol.Message{
				Type:       "chat",
				Text:       text,
				SenderName: c.Username,
			}

			delay := time.Duration(rand.Intn(300)) * time.Millisecond
			time.Sleep(delay)

			if err := c.WriteMessage(msgData); err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
				c.Connected = false
			}
		})

		p := tea.NewProgram(chatModel, tea.WithAltScreen())

		listenDone := make(chan struct{})
		go func() {
			c.Listen(p)
			close(listenDone)
		}()

		finalModel, err := p.Run()

		close(stopFake)

		if c.Conn != nil {
			c.Connected = false
			if closeErr := c.Conn.Close(); closeErr != nil {
				log.Printf("failed to close connection: %v", closeErr)
			}
		}

		<-listenDone

		if err == nil {
			if _, ok := finalModel.(ui.ChatModel); ok {
				fmt.Println("God loves the patient. Internet respects privacy.")
				return nil
			}
		}

		fmt.Printf("Reconnecting in %v...\n", c.Config.ReconnectDelay)
		utils.Spinner("Reconnecting", c.Config.ReconnectDelay)
	}
}
