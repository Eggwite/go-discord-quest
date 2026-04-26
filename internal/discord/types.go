package discord

const (
	OSWindows = "win32"
	OSDarwin  = "darwin"
	OSLinux   = "linux"
	OSAndroid = "android"
	OSIOS     = "ios"
)

type GameExecutable struct {
	IsLauncher  bool   `json:"is_launcher"`
	Name        string `json:"name"`
	OS          string `json:"os"`
	Filename    string `json:"filename,omitempty"`
	Path        string `json:"path,omitempty"`
	Segments    int    `json:"segments,omitempty"`
	IsRunning   bool   `json:"is_running,omitempty"`
	IsInstalled bool   `json:"is_installed,omitempty"`
}

type Game struct {
	UID         string           `json:"uid,omitempty"`
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Executables []GameExecutable `json:"executables"`
	Aliases     []string         `json:"aliases,omitempty"`
	Themes      []string         `json:"themes,omitempty"`
	IsRunning   bool             `json:"is_running,omitempty"`
	IsInstalled bool             `json:"is_installed,omitempty"`
}
