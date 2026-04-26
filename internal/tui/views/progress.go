// views/progress.go
package views

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Width(10)
	valStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#C0CAF5")).Bold(true)
	statusStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color("#e977b0")).
			Foreground(lipgloss.Color("#1A1B26"))
)

func RenderProgressCard(gameName, exePath string, elapsed, total time.Duration, bar progress.Model) string {
	percent := 0.0
	if total > 0 {
		percent = elapsed.Seconds() / total.Seconds()
	}

	// Use lipgloss.Center instead of Middle
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		statusStyle.Render(" ACTIVE "),
		" ",
		valStyle.Copy().Foreground(lipgloss.Color("#e977b0")).Render(gameName),
	)

	info := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s %s", labelStyle.Render("Process:"), valStyle.Render(filepath.Base(exePath))),
		fmt.Sprintf("%s %s", labelStyle.Render("Started:"), valStyle.Render(formatDuration(elapsed))),
		fmt.Sprintf("%s %s", labelStyle.Render("Goal:"), valStyle.Render(formatDuration(total))),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		info,
		"",
		bar.ViewAs(percent),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Italic(true).Render("Press 'q' to stop spoofing"),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Margin(1, 0).
		Render(content)
}

// formatDuration restored to fix the undefined error
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
