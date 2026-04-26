package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Soft Pink Pastel Theme
var (
	// Main accents - soft pinks
	Accent      = lipgloss.Color("#F472B6")
	AccentSoft  = lipgloss.Color("#F9A8D4")
	AccentMuted = lipgloss.Color("#FBCFE8")

	// UI elements
	Text      = lipgloss.Color("#2D2A2E")
	TextLight = lipgloss.Color("#6B6268")
	Muted     = lipgloss.Color("#A59BA2")

	// Status colors - pastel versions
	Success = lipgloss.Color("#A7F3D0")
	Danger  = lipgloss.Color("#FCA5A5")
	Warning = lipgloss.Color("#FDE047")

	// Backgrounds - soft cream/pink
	CardBorder  = lipgloss.Color("#FBCFE8")
	BannerBg    = lipgloss.Color("#FCE7F3")
	BannerRight = lipgloss.Color("#FDF2F8")
)

var (
	AppPadding = lipgloss.NewStyle().
			Padding(1, 2)

	CardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CardBorder).
			Padding(1, 1)

	InputCardStyle = CardStyle.Copy().
			BorderForeground(AccentSoft)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Soft pastel status bar
	StatusBarStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Style to make keys pop in the footer
	KeyStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// Selection highlight for lists
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Background(AccentMuted).
			Bold(true)

	// Pastel divider/border
	DividerStyle = lipgloss.NewStyle().
			Foreground(AccentMuted)
)

func Banner(width int, title, subtitle string) string {
	t := lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true).
		Render(title)

	s := lipgloss.NewStyle().
		Foreground(TextLight).
		Render(subtitle)

	content := lipgloss.JoinVertical(lipgloss.Left, t, s)

	b := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(Text)

	if width > 0 {
		b = b.Width(max(30, width))
	}

	return b.Render(content)
}

func StatChip(label, value string) string {
	l := lipgloss.NewStyle().
		Foreground(TextLight).
		Render(label)

	v := lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true).
		Render(value)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CardBorder).
		Padding(0, 2).
		Render(l + " " + v)
}

// New helper for styled separators
func Separator() string {
	return DividerStyle.Render(" • ")
}

// Progress bar in pastel pink
func PastelProgressBar(current, total int, width int) string {
	if total == 0 {
		total = 1
	}

	percent := float64(current) / float64(total)
	filled := int(float64(width) * percent)

	if filled > width {
		filled = width
	}

	bar := lipgloss.NewStyle().
		Background(AccentMuted).
		Width(width).
		Render(strings.Repeat(" ", width))

	filledPart := lipgloss.NewStyle().
		Background(Accent).
		Width(filled).
		Render(strings.Repeat(" ", filled))

	return lipgloss.JoinHorizontal(lipgloss.Top, filledPart, bar[filled:])
}
