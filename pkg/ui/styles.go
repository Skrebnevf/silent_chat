package ui

import "github.com/charmbracelet/lipgloss"

var (
	Nord0 = lipgloss.Color("#2E3440")
	Nord1 = lipgloss.Color("#3B4252")
	Nord2 = lipgloss.Color("#434C5E")
	Nord3 = lipgloss.Color("#4C566A")
)

var (
	Nord4 = lipgloss.Color("#D8DEE9")
	Nord5 = lipgloss.Color("#E5E9F0")
	Nord6 = lipgloss.Color("#ECEFF4")
)

var (
	Nord7  = lipgloss.Color("#8FBCBB")
	Nord8  = lipgloss.Color("#88C0D0")
	Nord9  = lipgloss.Color("#81A1C1")
	Nord10 = lipgloss.Color("#5E81AC")
)

var (
	Nord11 = lipgloss.Color("#BF616A")
	Nord12 = lipgloss.Color("#D08770")
	Nord13 = lipgloss.Color("#EBCB8B")
	Nord14 = lipgloss.Color("#A3BE8C")
	Nord15 = lipgloss.Color("#B48EAD")
)

var (
	ColorPrimary      = Nord9
	ColorPrimaryLight = Nord8
	ColorWhite        = Nord6
	ColorGray         = Nord4
	ColorGrayDark     = Nord3
	ColorError        = Nord11
	ColorText         = Nord6
	ColorBackground   = Nord0
	ColorSuccess      = Nord14
)

func AppBackgroundStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(ColorBackground).
		Foreground(ColorText).
		Width(width).
		Height(height)
}

func TitleStyle(width int) lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(Nord6).
		Background(Nord10).
		Padding(0, 1).
		Bold(true)

	if width > 0 {
		style = style.Width(width)
	}

	return style
}

func LabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Nord4).
		Bold(true)
}

func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)
}

func HelpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Nord3).
		Italic(true)
}

func MessageBoxStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Nord8).
		Width(width).
		Height(height).
		Padding(0, 1)
}

func InputBoxStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Nord9).
		Width(width).
		Padding(0, 1)
}

func SenderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Nord8).
		Bold(true)
}

func MessageTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Nord6)
}

func InputTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Nord6)
}
