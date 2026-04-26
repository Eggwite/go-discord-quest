<p>
  <h1 align="center">go-discord-quest (dqc)</h1>
</p>

https://github.com/user-attachments/assets/4eb8f725-6ff9-4b7f-b502-742d4a22f294

<p>
  <h4 align="center"> 📦 A single-file Windows TUI tool for automating Discord Quest completion.
</h1>
</p>

## 2-Step Quick Start

1. Go to the [Releases](../../releases) tab on this repository and download the latest `dqc-amd64.exe`.

2. Run `dqc-amd64.exe`. *No installation wizard or elevated Administrator permissions required.*


>### What it does:
> <img width="720" height="640" alt="image" align="center" src="https://github.com/user-attachments/assets/16d465d5-b2e9-4c65-9983-8920c59d175e" />



---

> [!WARNING]  
># Disclaimer
>
>This project is provided for educational and research purposes only. It is not intended to be used in violation of any platform’s terms of service.
>
>Use of this software may conflict with the terms set by **Discord**. You are solely responsible for ensuring that your use complies with the applicable terms, including the [**Discord Terms of Service**](https://discord.com/terms)
>
>I, the author, do not condone or encourage misuse of this tool or any activity that violates service agreements.

## Privacy

This tool does not collect, store, or transmit personal data.

- No telemetry
- No analytics
- No external tracking services

All operations are performed locally on your machine, **except for requests made directly to official Discord endpoints required for functionality.**

### Release Integrity

Each release includes a SHA256 checksum (`.sha256`) alongside the binary.

You can optionally verify the integrity of the downloaded executable:

```powershell
certutil -hashfile .\dqc-amd64.exe SHA256
```
---

## Development (DEVX)

### Prerequisites
- Go 1.26+ (current stable series)
- Windows 10/11 environment (required to build the Win32 stub and masquerading mechanisms)
- [Air](https://github.com/air-verse/air) (Optional, for live-reloading)

### Build from Source

```powershell 
# build.ps1 (you can run this included script)

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

Output: `dist/dqc.exe`

### Versioning & Releases

Releases are tagged using the format:

```
vX.Y.Z
```

Each tag triggers an automated GitHub Actions workflow which:

- Builds the Windows binary (`dqc.exe`)
- Generates a SHA256 checksum
- Publishes both as a GitHub Release

Only tagged commits are released. The `main` branch may contain unreleased changes.

### Logs & Debugging
The TUI maintains an in-memory ring-buffer of log entries (`LogInfo`, `LogWarning`, `LogError`). These logs can be dumped or extended using the internal `tui.Model` methods.
