package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	timerAccent = lipgloss.Color("#F472B6")
	timerText   = lipgloss.Color("#E2E2E6")
	timerMuted  = lipgloss.Color("#8B8D98")

	timerCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(timerAccent).
			Padding(1, 4).
			Align(lipgloss.Center)

	timerTagStyle = lipgloss.NewStyle().
			Foreground(timerAccent).
			Bold(true)

	timerChipStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(timerMuted).
			Padding(0, 2)

	timerChipLabelStyle = lipgloss.NewStyle().
				Foreground(timerMuted)

	timerChipValueStyle = lipgloss.NewStyle().
				Foreground(timerAccent).
				Bold(true)

	gameNameStyle = lipgloss.NewStyle().
			Foreground(timerText).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(timerMuted)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(timerAccent).
			Bold(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(timerMuted)
)

func RenderTimerCard(gameName string, duration time.Duration) string {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	tag := timerTagStyle.Render("QUEST TIMER")

	game := gameNameStyle.Render(gameName)

	durationChip := timerChipStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Center,
			timerChipLabelStyle.Render("DURATION "),
			timerChipValueStyle.Render(timeStr),
		),
	)

	rangeNote := mutedStyle.Render(
		fmt.Sprintf("(%s – %s range)", formatDurationMin(5*time.Minute), formatDurationMin(60*time.Minute)),
	)

	divider := dividerStyle.Render("─────────────────────")

	controls := lipgloss.JoinVertical(lipgloss.Center,
		fmt.Sprintf("%s  %s", keyHintStyle.Render("↑ / ↓"), mutedStyle.Render("adjust by 30s")),
		fmt.Sprintf("%s  %s", keyHintStyle.Render("enter"), mutedStyle.Render("confirm & spoof")),
		fmt.Sprintf("%s  %s", keyHintStyle.Render("esc"), mutedStyle.Render("back to search")),
	)

	content := lipgloss.JoinVertical(lipgloss.Center,
		tag,
		"",
		game,
		"",
		durationChip,
		rangeNote,
		"",
		divider,
		"",
		controls,
	)

	return timerCardStyle.Render(content)
}

func formatDurationMin(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if s > 0 {
		return fmt.Sprintf("%d:%02d", m, s)
	}
	return fmt.Sprintf("%d:00", m)
}
