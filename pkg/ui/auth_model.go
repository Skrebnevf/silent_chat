package ui

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type AuthData struct {
	Host     string
	Port     string
	Password string
	Username string
}

type AuthModel struct {
	step      int
	inputs    []textinput.Model
	err       error
	completed bool
	authData  *AuthData
	width     int
	height    int
}

func NewAuthModel() AuthModel {
	host := textinput.New()
	host.Placeholder = "192.168.0.1"
	host.Width = 40
	host.Focus()

	port := textinput.New()
	port.Width = 40
	port.Placeholder = "4000"

	pass := textinput.New()
	pass.Placeholder = "********"
	pass.EchoMode = textinput.EchoPassword
	pass.Width = 40
	pass.EchoCharacter = '*'

	user := textinput.New()
	user.Width = 40
	user.Placeholder = "username"

	return AuthModel{
		inputs: []textinput.Model{host, port, pass, user},
	}
}

func (m AuthModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.step < 3 {
				m.step++
				m.inputs[m.step].Focus()
			}
		case "shift+tab":
			if m.step > 0 {
				m.step--
				m.inputs[m.step].Focus()
			}
		case "enter":
			if m.step < 3 {
				m.step++
				m.inputs[m.step].Focus()
			} else {
				h, p, pw, u := m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value(), m.inputs[3].Value()
				if h == "" || p == "" || pw == "" || u == "" {
					m.err = fmt.Errorf("all fields required")
					return m, nil
				}
				if net.ParseIP(h) == nil && !isValidDomain(h) {
					m.err = fmt.Errorf("invalid host: must be a valid IP address or domain")
					return m, nil
				}
				portNum, err := strconv.Atoi(p)
				if err != nil || portNum < 1 || portNum > 65535 {
					m.err = fmt.Errorf("invalid port: must be a number between 1 and 65535")
					return m, nil
				}
				m.completed = true
				m.authData = &AuthData{Host: h, Port: p, Password: pw, Username: u}
				return m, tea.Quit
			}
		}
	}

	for i := range m.inputs {
		if i == m.step {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
		m.inputs[i], _ = m.inputs[i].Update(msg)
	}
	return m, nil
}

func (m AuthModel) View() string {
	labels := []string{"Host", "Port", "Password", "Username"}
	var s strings.Builder

	s.WriteString(GetASCIIArt() + "\n")
	// s.WriteString(GetWelcomeText() + "\n\n")
	s.WriteString(GetSeparator(60) + "\n\n")

	title := TitleStyle(0).Render(" Welcome. Please add connection settings. ")
	s.WriteString(title + "\n\n")

	for i, label := range labels {
		s.WriteString(LabelStyle().Render(label) + "\n")
		s.WriteString(m.inputs[i].View() + "\n\n")
	}

	if m.err != nil {
		s.WriteString(ErrorStyle().Render("Error: "+m.err.Error()) + "\n")
	}

	s.WriteString(
		HelpStyle().Render("Press Enter to continue or Ctrl+C to quit"),
	)

	return s.String()
}

func (m AuthModel) GetAuthData() (*AuthData, error) {
	if !m.completed {
		return nil, fmt.Errorf("authentication cancelled")
	}
	return m.authData, nil
}

func isValidDomain(domain string) bool {
	if strings.Contains(domain, " ") || !strings.Contains(domain, ".") {
		return false
	}
	return len(domain) > 0
}
