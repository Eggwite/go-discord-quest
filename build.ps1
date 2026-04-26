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