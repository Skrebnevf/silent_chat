package ui

import "github.com/charmbracelet/lipgloss"

const silentChatASCII = `
  ____ ___ _     _____ _   _ _____    ____ _   _    _  _____ 
 / ___|_ _| |   | ____| \ | |_   _|  / ___| | | |  / \|_   _|
 \___ \| || |   |  _| |  \| | | |   | |   | |_| | / _ \ | |  
  ___) | || |___| |___| |\  | | |   | |___|  _  |/ ___ \| |  
 |____/___|_____|_____|_| \_| |_|    \____|_| |_/_/   \_\_|  
`

// const welcomeText = “
func GetASCIIArt() string {
	style := lipgloss.NewStyle().
		Foreground(Nord8).
		Bold(true).
		Align(lipgloss.Center)

	return style.Render(silentChatASCII)
}

// func GetWelcomeText() string {
// 	style := lipgloss.NewStyle().
// 		Foreground(Nord7).
// 		Italic(true).
// 		Align(lipgloss.Center)

// 	return style.Render(welcomeText)
// }

func GetSeparator(width int) string {
	if width <= 0 {
		width = 60
	}

	style := lipgloss.NewStyle().
		Foreground(Nord3)

	separator := ""
	for i := 0; i < width; i++ {
		separator += "─"
	}

	return style.Render(separator)
}
