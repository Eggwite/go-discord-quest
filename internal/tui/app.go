package tui

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Eggwite/go-discord-quest/internal/discord"
	"github.com/Eggwite/go-discord-quest/internal/runner"
	"github.com/Eggwite/go-discord-quest/internal/search"
	"github.com/Eggwite/go-discord-quest/internal/tui/views"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const QuestDuration = 15 * time.Minute

type AppState int

const (
	StateLoading AppState = iota
	StateSearch
	StateRunning
	StateDone
)

type LogType string

const (
	LogInfo    LogType = "info"
	LogError   LogType = "error"
	LogWarning LogType = "warning"
	LogDebug   LogType = "debug"
)

type LogEntry struct {
	Type      LogType
	Message   string
	Timestamp time.Time
}

type gamesLoadedMsg struct {
	games []discord.Game
	trace []string
	err   error
}
type processExitedMsg struct {
	err error
}
type runnerStartedMsg struct {
	runner *runner.Runner
	err    error
}

type tickMsg time.Time

type Model struct {
	state AppState

	spinner spinner.Model
	input   textinput.Model
	list    list.Model
	bar     progress.Model
	help    help.Model
	vp      viewport.Model
	keys    KeyMap

	showLogs bool

	allGames []discord.Game
	filtered []discord.Game
	selected discord.Game

	runner  *runner.Runner
	started time.Time
	elapsed time.Duration

	width  int
	height int

	loadingTrace []string
	logs         []LogEntry
}

func New() *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(Accent)

	input := views.NewSearchInput()
	lst := views.NewGameList(nil)
	bar := progress.New(progress.WithScaledGradient("#5865F2", "#57F287"))

	hp := help.New()
	vp := viewport.New(80, 10)
	vp.Style = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(Muted)

	return &Model{
		state:        StateLoading,
		spinner:      sp,
		input:        input,
		list:         lst,
		bar:          bar,
		help:         hp,
		vp:           vp,
		keys:         DefaultKeyMap(),
		loadingTrace: []string{"* Fetching from GitHub mirror...", "* Fetching from Discord API...", "* Using bundled game list"},
		logs:         make([]LogEntry, 0, 32),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadGamesCmd())
}

func loadGamesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		games, trace, err := discord.FetchDetectableWithTrace(ctx)
		return gamesLoadedMsg{games: games, trace: trace, err: err}
	}
}

func startRunnerCmd(game discord.Game) tea.Cmd {
	return func() tea.Msg {
		r, err := runner.Start(game)
		return runnerStartedMsg{runner: r, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-8, m.height-18)
		m.help.Width = msg.Width
		listHeight := max(8, msg.Height-8)
		if m.showLogs {
			listHeight = max(8, msg.Height-18)
			m.vp.Width = max(40, msg.Width-4)
			m.vp.Height = 8
		}
		m.list.SetSize(max(30, msg.Width-4), listHeight)

	case spinner.TickMsg:
		if m.state == StateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case gamesLoadedMsg:
		if msg.err != nil {
			m.appendLog(LogError, fmt.Sprintf("failed to load games: %v", msg.err))
			m.allGames = []discord.Game{}
			m.filtered = []discord.Game{}
			m.list.SetItems(nil)
		} else {
			m.appendLog(LogInfo, fmt.Sprintf("loaded %d games", len(msg.games)))
			m.allGames = msg.games
			m.filtered = search.Search(m.allGames, "")
			m.list.SetItems(views.ToItems(m.filtered))
		}
		if len(msg.trace) > 0 {
			m.loadingTrace = msg.trace
		}
		m.state = StateSearch
		return m, nil

	case runnerStartedMsg:
		if msg.err != nil {
			m.appendLog(LogError, fmt.Sprintf("failed to start runner: %v", msg.err))
			m.state = StateSearch
			return m, nil
		}
		m.runner = msg.runner
		m.started = time.Now()
		m.elapsed = 0
		m.appendLog(LogInfo, "runner started")
		return m, tickCmd()

	case tickMsg:
		if m.state != StateRunning || m.runner == nil {
			return m, nil
		}

		// Check if process is still running
		if !m.runner.IsRunning() {
			m.appendLog(LogInfo, "Game window was closed manually")
			m.state = StateSearch
			m.runner = nil
			return m, nil
		}

		m.elapsed = time.Since(m.started)
		if m.elapsed >= QuestDuration {
			_ = m.runner.Stop()
			m.runner = nil
			m.state = StateDone
			m.appendLog(LogInfo, "quest completed")
			return m, nil
		}
		return m, tickCmd()

	case tea.KeyMsg:
		var cmds []tea.Cmd

		switch msg.String() {
		case "ctrl+l":
			m.showLogs = !m.showLogs
			// If we are opening logs, ensure the viewport is scrolled to the bottom
			if m.showLogs {
				m.vp.GotoBottom()
			}
			// Force a resize/repaint sync
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}

		case "ctrl+c", "q": // Standard exit
			return m, tea.Quit
		}

		// IMPORTANT: Pass the key event to the viewport when logs are visible
		// But don't swallow the event - let the viewport handle it if logs are showing
		if m.showLogs {
			// Also pass the key to the viewport to handle scrolling
			var vpCmd tea.Cmd
			m.vp, vpCmd = m.vp.Update(msg)
			// Only return the viewport command if we're not also handling other commands
			if len(cmds) == 0 {
				return m, vpCmd
			}
			cmds = append(cmds, vpCmd)
		}

		switch m.state {
		case StateSearch:
			switch msg.String() {
			case "esc", "ctrl+c":
				return m, tea.Quit
			case "enter":
				game, ok := views.SelectedGame(m.list)
				if !ok {
					return m, nil
				}
				m.selected = game
				m.state = StateRunning
				m.elapsed = 0
				return m, startRunnerCmd(game)
			}

			var inputCmd tea.Cmd
			m.input, inputCmd = m.input.Update(msg)
			cmds = append(cmds, inputCmd)

			m.filtered = search.Search(m.allGames, m.input.Value())
			m.list.SetItems(views.ToItems(m.filtered))

			var listCmd tea.Cmd
			m.list, listCmd = m.list.Update(msg)
			cmds = append(cmds, listCmd)

			return m, tea.Batch(cmds...)

		case StateRunning:
			if msg.String() == "q" || msg.String() == "esc" {
				if m.runner != nil {
					_ = m.runner.Stop()
					m.runner = nil
				}
				m.state = StateSearch
				m.appendLog(LogWarning, "runner stopped by user")
			}

		case StateDone:
			m.state = StateSearch
			m.elapsed = 0
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *Model) View() string {
	docWidth := m.width - 4
	if docWidth < 0 {
		docWidth = 80
	}

	header := Banner(docWidth, " DISCORD QUEST COMPLETER ", " Spoofing background processes for fun and profit ")

	var body string
	switch m.state {
	case StateLoading:
		body = m.spinner.View() + " Initialising..."

	case StateSearch:
		stats := lipgloss.JoinHorizontal(lipgloss.Left,
			StatChip("LOADED", fmt.Sprintf("%d", len(m.allGames))),
			" ",
			StatChip("MATCHED", fmt.Sprintf("%d", len(m.filtered))),
		)
		input := InputCardStyle.Width(docWidth - 2).Render(m.input.View())

		var content string
		if m.showLogs {
			// When showing logs, display the log viewport
			// Adjust list height to make room for logs
			m.list.SetSize(docWidth-4, m.height-26) // Reduced height to make room for logs
			logsView := lipgloss.NewStyle().
				Width(docWidth).
				MarginTop(1).
				Render(m.vp.View())
			listView := CardStyle.Width(docWidth - 4).Render(m.list.View())
			content = lipgloss.JoinVertical(lipgloss.Left, input, listView, logsView)
		} else {
			m.list.SetSize(docWidth-4, m.height-15)
			listView := CardStyle.Width(docWidth - 2).Render(m.list.View())
			content = lipgloss.JoinVertical(lipgloss.Left, input, listView)
		}

		body = lipgloss.JoinVertical(lipgloss.Left, stats, content)

	case StateRunning:
		// FIX: Nil guard to prevent panic before runner starts
		exePath := "Launching..."
		if m.runner != nil {
			exePath = m.runner.ExePath
		}
		body = views.RenderProgressCard(m.selected.Name, exePath, m.elapsed, QuestDuration, m.bar)

	case StateDone:
		body = lipgloss.NewStyle().Foreground(Success).Bold(true).Render("Quest completed! Press any key to return.")
	}

	helpView := m.help.View(m.keys)
	footer := StatusBarStyle.Width(m.width).Render(helpView)

	fullUI := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"\n",
		body,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		AppPadding.Render(fullUI),
		lipgloss.NewStyle().MarginTop(1).Render(footer),
	)
}

func (m *Model) KillRunner() {
	if m.runner != nil {
		_ = m.runner.Stop()
		m.runner = nil
	}
}

func (m *Model) appendLog(level LogType, message string) {
	m.logs = append(m.logs, LogEntry{Type: level, Message: message, Timestamp: time.Now()})
	if len(m.logs) > 200 {
		m.logs = m.logs[len(m.logs)-200:]
	}

	var sb strings.Builder
	for _, lg := range m.logs {
		color := Muted
		switch lg.Type {
		case LogError:
			color = lipgloss.Color("#ED4245")
		case LogWarning:
			color = lipgloss.Color("#FEE75C")
		case LogInfo:
			color = lipgloss.Color("#57F287")
		}
		prefix := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("[%s]", lg.Type))
		sb.WriteString(fmt.Sprintf("%s %s %s\n", lg.Timestamp.Format("15:04:05"), prefix, lg.Message))
	}
	m.vp.SetContent(sb.String())
	m.vp.GotoBottom()
}

func firstWindowsExe(game discord.Game) string {
	for _, exe := range game.Executables {
		if exe.OS == discord.OSWindows && !exe.IsLauncher {
			if out := normalizeExeDisplay(exe); out != "" {
				return out
			}
		}
	}
	for _, exe := range game.Executables {
		if exe.OS == discord.OSWindows {
			if out := normalizeExeDisplay(exe); out != "" {
				return out
			}
		}
	}
	return ""
}

func normalizeExeDisplay(exe discord.GameExecutable) string {
	name := strings.TrimSpace(exe.Filename)
	if name == "" {
		name = strings.TrimSpace(exe.Name)
	}
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.ReplaceAll(name, ">", "")
	return path.Base(name)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
