package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ChatModel struct {
	messages     []chatMsg
	input        textinput.Model
	username     string
	onSend       func(string)
	width        int
	height       int
	scrollOffset int
}

type chatMsg struct {
	Sender string
	Text   string
}

type NewChatMsg struct {
	Sender string
	Text   string
}

func NewChatModel(username string, onSend func(string)) ChatModel {
	ti := textinput.New()
	ti.Placeholder = "Type a message... for exit type /quit or CTRL+C to exit"
	ti.Focus()
	ti.CharLimit = 500

	return ChatModel{
		messages:     make([]chatMsg, 0, 100),
		input:        ti,
		username:     username,
		onSend:       onSend,
		width:        80,
		height:       24,
		scrollOffset: 0,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 10
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
			return m, nil
		case "down":
			availableHeight := m.height - 8
			if availableHeight < 1 {
				availableHeight = 1
			}
			maxScroll := len(m.messages) - availableHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollOffset < maxScroll {
				m.scrollOffset++
			}
			return m, nil
		case "enter":
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				break
			}
			if text == "/quit" {
				return m, tea.Quit
			}
			m.messages = append(m.messages, chatMsg{Sender: m.username, Text: text})
			m.input.SetValue("")
			m.scrollToBottom()
			if m.onSend != nil {
				m.onSend(text)
			}
		}

	case NewChatMsg:
		m.messages = append(m.messages, chatMsg{Sender: msg.Sender, Text: msg.Text})
		m.scrollToBottom()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *ChatModel) scrollToBottom() {
	availableHeight := m.height - 8
	if availableHeight < 1 {
		availableHeight = 1
	}
	maxScroll := len(m.messages) - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	m.scrollOffset = maxScroll
}

func (m ChatModel) View() string {
	var s strings.Builder

	title := TitleStyle(
		m.width,
	).Render(fmt.Sprintf("Username — %s", m.username))
	s.WriteString(title + "\n")

	availableHeight := m.height - 8
	if availableHeight < 1 {
		availableHeight = 1
	}

	startIdx := m.scrollOffset
	endIdx := m.scrollOffset + availableHeight
	if endIdx > len(m.messages) {
		endIdx = len(m.messages)
	}

	var messagesContent strings.Builder
	messageLines := 0
	for i := startIdx; i < endIdx; i++ {
		msg := m.messages[i]
		sender := SenderStyle().Render(msg.Sender + ":")
		text := MessageTextStyle().Render(" " + msg.Text)
		messagesContent.WriteString(sender + text + "\n")
		messageLines++
	}

	for i := messageLines; i < availableHeight; i++ {
		messagesContent.WriteString("\n")
	}

	messagesBox := MessageBoxStyle(m.width-4, availableHeight).
		Render(strings.TrimRight(messagesContent.String(), "\n"))
	s.WriteString(messagesBox + "\n")

	scrollInfo := ""
	if len(m.messages) > availableHeight {
		scrollInfo = "↑↓ to scroll"
	}
	s.WriteString(HelpStyle().Render(scrollInfo) + "\n")

	inputBox := InputBoxStyle(m.width - 4).Render(m.input.View())
	s.WriteString(inputBox)

	return s.String()
}
