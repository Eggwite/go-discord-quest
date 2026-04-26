package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/Eggwite/go-discord-quest/internal/discord"
)

type Runner struct {
	Game      discord.Game
	ExePath   string
	GameDir   string
	cmd       *exec.Cmd
	done      chan error
	stopCheck chan struct{}
}

func Start(game discord.Game) (*Runner, error) {
	exePath, gameDir := gameExePath(game)
	if err := os.MkdirAll(filepath.Dir(exePath), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	if len(stubBytes) == 0 {
		return nil, fmt.Errorf("embedded stub is empty")
	}
	if err := os.WriteFile(exePath, stubBytes, 0o755); err != nil {
		return nil, fmt.Errorf("write stub: %w", err)
	}

	cmd := exec.Command(exePath, "--title", game.Name)
	cmd.Dir = filepath.Dir(exePath)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("spawn: %w", err)
	}

	runner := &Runner{
		Game:      game,
		ExePath:   exePath,
		GameDir:   gameDir,
		cmd:       cmd,
		done:      make(chan error, 1),
		stopCheck: make(chan struct{}),
	}

	go runner.monitorProcess()
	return runner, nil
}

func (r *Runner) monitorProcess() {
	if r.cmd == nil || r.cmd.Process == nil {
		return
	}

	err := r.cmd.Wait()

	select {
	case <-r.stopCheck:
		return
	case r.done <- err:
	}
}

func (r *Runner) WaitForExit() <-chan error {
	return r.done
}

func (r *Runner) IsRunning() bool {
	if r.cmd == nil || r.cmd.Process == nil {
		return false
	}

	if runtime.GOOS == "windows" {
		select {
		case <-r.done:
			return false
		default:
			return true
		}
	}

	err := r.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func (r *Runner) Stop() error {
	if r == nil {
		return nil
	}

	close(r.stopCheck)

	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Signal(os.Interrupt)
		_ = r.cmd.Process.Kill()
	}

	if r.ExePath == "" {
		return nil
	}
	if r.GameDir != "" {
		return os.RemoveAll(r.GameDir)
	}
	return os.RemoveAll(filepath.Dir(r.ExePath))
}

// All the helper functions below need to be in the same file

func baseDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func stubTemplatePath() string {
	return filepath.Join(baseDir(), "data", "stub.exe")
}

func gameExePath(game discord.Game) (string, string) {
	gameDir := filepath.Join(gameStorageRoot(), sanitizePathSegment(game.ID))
	relDir, exe := exeLayout(game)
	return filepath.Join(gameDir, relDir, exe), gameDir
}

func gameStorageRoot() string {
	if root := strings.TrimSpace(os.Getenv("DQC_GAMES_ROOT")); root != "" {
		return root
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(baseDir(), "games")
	}
	return filepath.Join(home, "Documents", "DiscordQuestGames")
}

func exeLayout(game discord.Game) (string, string) {
	if exe, ok := preferredWindowsExecutable(game.Executables); ok {
		relDir, name := normalizeExecutableLayout(exe)
		if name != "" {
			if !strings.HasSuffix(strings.ToLower(name), ".exe") {
				name += ".exe"
			}
			return relDir, name
		}
	}

	safe := sanitizeFileName(game.Name)
	if safe == "" {
		safe = "game"
	}
	return "", safe + ".exe"
}

func preferredWindowsExecutable(executables []discord.GameExecutable) (discord.GameExecutable, bool) {
	for _, exe := range executables {
		if exe.OS == discord.OSWindows && !exe.IsLauncher {
			return exe, true
		}
	}
	for _, exe := range executables {
		if exe.OS == discord.OSWindows {
			return exe, true
		}
	}
	return discord.GameExecutable{}, false
}

func normalizeExecutableLayout(exe discord.GameExecutable) (string, string) {
	rawName := strings.TrimSpace(exe.Name)
	rawFile := strings.TrimSpace(exe.Filename)
	rawPath := strings.TrimSpace(exe.Path)

	if rawFile == "" {
		rawFile = path.Base(strings.ReplaceAll(rawName, "\\", "/"))
	}

	if rawPath == "" {
		rawName = strings.ReplaceAll(rawName, "\\", "/")
		rawName = strings.TrimSpace(strings.ReplaceAll(rawName, ">", ""))
		rawPath = path.Dir(rawName)
		if rawPath == "." {
			rawPath = ""
		}
	}

	return sanitizePath(rawPath), sanitizeFileName(rawFile)
}

func sanitizePath(pathLike string) string {
	pathLike = strings.TrimSpace(pathLike)
	pathLike = strings.ReplaceAll(pathLike, "\\", "/")
	pathLike = strings.ReplaceAll(pathLike, ">", "")
	pathLike = strings.TrimPrefix(pathLike, "/")

	parts := strings.Split(pathLike, "/")
	safe := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." || part == ".." {
			continue
		}
		part = sanitizePathSegment(part)
		if part != "" {
			safe = append(safe, part)
		}
	}
	if len(safe) == 0 {
		return ""
	}
	return filepath.Join(safe...)
}

func sanitizePathSegment(s string) string {
	if len(s) >= 2 && s[1] == ':' {
		s = s[:1]
	}
	return sanitizeFileName(s)
}

func sanitizeFileName(s string) string {
	replacer := strings.NewReplacer(
		"\\", "",
		"/", "",
		":", "",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	return strings.TrimSpace(replacer.Replace(s))
}
