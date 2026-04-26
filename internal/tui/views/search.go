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

var (
	accent    = lipgloss.Color("#7D56F4")
	highlight = lipgloss.Color("#df93df")
	bg        = lipgloss.Color("#f5e5f1")
	subtle    = lipgloss.Color("#89566a")
	text      = lipgloss.Color("#C0CAF5")

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

type GameItem struct {
	Game discord.Game
	Desc string
}

func (g GameItem) Title() string       { return g.Game.Name }
func (g GameItem) FilterValue() string { return g.Game.Name }
func (g GameItem) Description() string { return g.Desc }

// NewGameItem caches the display label for the executable to save CPU cycles during render
func NewGameItem(g discord.Game) GameItem {
	desc := "No win32 executable"
	for _, exe := range g.Executables {
		if exe.OS == discord.OSWindows && !exe.IsLauncher {
			name := strings.TrimSpace(exe.Filename)
			if name == "" {
				name = strings.TrimSpace(exe.Name)
			}
			desc = path.Base(strings.ReplaceAll(name, "\\", "/"))
			break
		}
	}
	return GameItem{Game: g, Desc: desc}
}

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
	l.KeyMap.Quit.SetKeys()
	l.KeyMap.ForceQuit.SetKeys()
	return l
}

func ToItems(games []discord.Game) []list.Item {
	items := make([]list.Item, len(games))
	for i, g := range games {
		items[i] = NewGameItem(g)
	}
	return items
}

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

	availWidth := m.Width() - 2
	title := i.Title()
	exe := i.Description()

	baseStyle := lipgloss.NewStyle().Background(lipgloss.Color(""))
	titleStyle := baseStyle.Copy().PaddingLeft(1)
	exeStyle := baseStyle.Copy().Foreground(subtle).Italic(true)

	if index == m.Index() {
		titleStyle = titleStyle.
			Foreground(highlight).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(highlight).
			PaddingLeft(0)

		exeStyle = exeStyle.Foreground(accent).Bold(false)
	} else {
		titleStyle = titleStyle.Foreground(text)
	}

	renderedTitle := titleStyle.Render(title)
	renderedExe := exeStyle.Render(exe)

	paddingCount := availWidth - lipgloss.Width(renderedTitle) - lipgloss.Width(renderedExe)
	if paddingCount < 0 {
		paddingCount = 0
	}
	gap := strings.Repeat(" ", paddingCount)

	fmt.Fprintf(w, "%s%s%s", renderedTitle, gap, renderedExe)
}
