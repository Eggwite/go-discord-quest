# go-discord-quest (dqc)

A single-binary Windows CLI/TUI tool for automating Discord Quest completion without any heavy GUI frameworks or runtime dependencies.

> [!WARNING]  
># Disclaimer
>
>This project is provided for educational and research purposes only. It is not intended to be used in violation of any platform’s terms of service.
>
>Use of this software may conflict with the terms set by **Discord**. You are solely responsible for ensuring that your use complies with the applicable terms, including the **Discord Terms of Service**:
>https://discord.com/terms
>
>I do not condone or encourage misuse of this tool or any activity that violates service agreements.

## Features
- **Zero GUI Overhead**: Written entirely in Go with a terminal user interface (TUI) powered by Bubble Tea and Lipgloss.
- **Smart Game Search**: Fuzzy search ranks games by title, aliases, and executable names for quick selection.
- **Lightweight Stub Process**: Runs a minimal, size-optimized ~ 2.5MB Windows application, masquerading as the target game. This process sits quietly without interrupting your workflow.
- **Deterministic Builds (CI)**: Release binaries are built in a clean CI environment with consistent flags to minimise environmental differences.

## Installation & Usage (For Users)

### 1. Download the Latest Release
Go to the [Releases](../../releases) tab on this repository and download the latest `dqc.exe`.

### 2. Run the Application
Run `dqc.exe`. No installation wizard or elevated Administrator permissions are required!

```shell
# From PowerShell or CMD
.\dqc.exe
```

When you select a game, the tool will:
1. Copy the internal lightweight game stub into `%USERPROFILE%\\Documents\\DiscordQuestGames\\<app_id>\\...`.
2. Launch the stub as a new windowed application.
3. Automatically kill the stub when the 15-minute quest duration is completed, when you close the window, or you press `q` / `Esc`.

## Release Integrity

Each release includes a SHA256 checksum (`.sha256`) alongside the binary.

You can verify the integrity of the downloaded executable:

```powershell
certutil -hashfile .\dqc.exe SHA256
```
---

## Developer Setup (DEVX)

### Prerequisites
- Go 1.26+ (current stable series)
- Windows 10/11 environment (required to build the Win32 stub and masquerading mechanisms)
- [Air](https://github.com/air-verse/air) (Optional, for live-reloading)

### Build from Source

This project uses standard Go tooling and does not require `make` or any Unix-like environment.

#### PowerShell (Recommended)

```powershell 
# build.ps1

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build `
	-ldflags="-H windowsgui -s -w" -trimpath
-o internal/runner/assets/stub.exe `
	./stub

# Build main binary
go build `
	-ldflags="-s -w" -trimpath
-o dist/dqc.exe `
	./cmd/dqc
```
`-trimpath` removes local filesystem paths from the binary, improving build reproducibility.
The final artifact will be located at dist/dqc.exe.

### Live Reload / Debugging (Using Air)
To improve the development experience with hot-reloading:

1. Install Air:
   ```shell
   go install github.com/air-verse/air@latest
   ```
2. Start Air:
   ```shell
   air
   ```
   Air will automatically rebuild the stub and restart the TUI whenever you save a `.go` file.
## Versioning & Releases

Releases are tagged using the format:

```
vX.Y.Z
```

Each tag triggers an automated GitHub Actions workflow which:

- Builds the Windows binary (`dqc.exe`)
- Generates a SHA256 checksum
- Publishes both as a GitHub Release

Example:
- `v1.2.0` → produces `dqc.exe` and `dqc-amd64.sha256`

Only tagged commits are released. The `main` branch may contain unreleased changes.

## Architecture

- **`internal/tui/`**: Contains the Bubble Tea state machine, views (progress, search), and styling.
- **`internal/runner/`**: Manages deploying the embedded executable into a local folder and tracking the running process.
- **`internal/discord/`**: Defines the target typings and Discord API for mapping games correctly.
- **`internal/search/`**: Provides weighted fuzzy sorting across multiple game metadata fields to prioritise exact game matches.
- **`stub/`**: A pure `x/sys/windows` Win32 implementation handling standard window creation.

## Logs & Debugging
The TUI maintains an in-memory ring-buffer of log entries (`LogInfo`, `LogWarning`, `LogError`). These logs can be dumped or extended using the internal `tui.Model` methods.

## Privacy

This tool does not collect, store, or transmit personal data.

- No telemetry
- No analytics
- No external tracking services

All operations are performed locally on your machine, **except for requests made directly to official Discord endpoints required for functionality.**
