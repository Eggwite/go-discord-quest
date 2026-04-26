package views

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/Eggwite/go-discord-quest/internal/discord"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- UI Theme Palette ---
var (
	accent    = lipgloss.Color("#7D56F4")
	highlight = lipgloss.Color("#df93df")
	bg        = lipgloss.Color("#f5e5f1")
	subtle    = lipgloss.Color("#89566a")
	text      = lipgloss.Color("#C0CAF5") // Soft White

	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2).
			Background(bg)

	headerStyle = lipgloss.NewStyle().
			Foreground(bg).
			Background(highlight).
			Padding(0, 1).
			Bold(true)
)

// --- GameItem Implementation ---

type GameItem struct {
	Game discord.Game
}

func (g GameItem) Title() string       { return g.Game.Name }
func (g GameItem) FilterValue() string { return fmt.Sprintf("%s %v", g.Game.Name, g.Game.Aliases) }

func (g GameItem) Description() string {
	for _, exe := range g.Game.Executables {
		if exe.OS == discord.OSWindows && !exe.IsLauncher {
			if name := normalizeExeLabel(exe); name != "" {
				return name
			}
		}
	}
	return "No win32 executable"
}

func normalizeExeLabel(exe discord.GameExecutable) string {
	name := strings.TrimSpace(exe.Filename)
	if name == "" {
		name = strings.TrimSpace(exe.Name)
	}
	name = strings.ReplaceAll(name, "\\", "/")
	return path.Base(name)
}

// --- Component Constructors ---

func NewSearchInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Find a game..."

	input.Focus()
	input.CharLimit = 64
	input.TextStyle = lipgloss.NewStyle().Foreground(text)
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(subtle)
	return input
}

func NewGameList(items []list.Item) list.Model {
	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "DISCORD QUESTS"
	l.Styles.Title = headerStyle

	l.SetShowPagination(true)
	l.SetShowStatusBar(false)

	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return l
}

// ToItems converts the discord.Game slice into list.Item interface compatible types.
func ToItems(games []discord.Game) []list.Item {
	items := make([]list.Item, 0, len(games))
	for _, g := range games {
		items = append(items, GameItem{Game: g})
	}
	return items
}

// SelectedGame extracts the underlying discord.Game from the list's current selection.
func SelectedGame(l list.Model) (discord.Game, bool) {
	item, ok := l.SelectedItem().(GameItem)
	if !ok {
		return discord.Game{}, false
	}
	return item.Game, true
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(GameItem)
	if !ok || listItem == nil {
		return
	}

	// Calculate width for justification
	availWidth := m.Width() - 2

	title := i.Title()
	exe := i.Description()

	// 1. Force transparent background for all base styles
	baseStyle := lipgloss.NewStyle().Background(lipgloss.Color(""))

	titleStyle := baseStyle.Copy().PaddingLeft(1)
	exeStyle := baseStyle.Copy().Foreground(subtle).Italic(true)

	// 2. Selection Styling
	if index == m.Index() {
		titleStyle = titleStyle.
			Foreground(highlight).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(highlight).
			PaddingLeft(0) // Align for border

		exeStyle = exeStyle.Foreground(accent).Bold(false)
	} else {
		titleStyle = titleStyle.Foreground(text)
	}

	// 3. Render strings
	renderedTitle := titleStyle.Render(title)
	renderedExe := exeStyle.Render(exe)

	// 4. Manual Justification with standard spaces
	// Using strings.Repeat ensures no hidden style-shading is applied to the gap
	paddingCount := availWidth - lipgloss.Width(renderedTitle) - lipgloss.Width(renderedExe)
	if paddingCount < 0 {
		paddingCount = 0
	}
	gap := strings.Repeat(" ", paddingCount)

	// Assembly
	fmt.Fprintf(w, "%s%s%s", renderedTitle, gap, renderedExe)
}
