package discord

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	discordURL = "https://discord.com/api/applications/detectable"
)

//go:embed assets/gamelist.json
var bundledList []byte

func FetchDetectable(ctx context.Context) ([]Game, error) {
	games, _, err := FetchDetectableWithTrace(ctx)
	return games, err
}

func FetchDetectableWithTrace(ctx context.Context) ([]Game, []string, error) {
	trace := make([]string, 0, 4)

	trace = append(trace, "* Fetching from Discord API...")
	if games, err := fetchURL(ctx, discordURL); err == nil {
		trace = append(trace, "  + Done")
		return filterWindows(games), trace, nil
	} else {
		trace = append(trace, fmt.Sprintf("  - Failed (%v), using bundled list", err))
	}

	trace = append(trace, "* Using bundled game list")
	games, err := parseBundled()
	if err != nil {
		return nil, trace, err
	}

	return filterWindows(games), trace, nil
}

func fetchURL(ctx context.Context, url string) ([]Game, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	games, err := parseGames(body)
	if err != nil {
		return nil, err
	}
	if len(games) == 0 {
		return nil, errors.New("empty game list")
	}
	return games, nil
}

func parseBundled() ([]Game, error) {
	games, err := parseGames(bundledList)
	if err != nil {
		return nil, err
	}
	if len(games) == 0 {
		return nil, errors.New("bundled list is empty")
	}
	return games, nil
}

func parseGames(raw []byte) ([]Game, error) {
	var games []Game
	if err := json.Unmarshal(raw, &games); err != nil {
		return nil, err
	}
	return games, nil
}

func filterWindows(games []Game) []Game {
	out := make([]Game, 0, len(games))
	for _, g := range games {
		for _, exe := range g.Executables {
			if exe.OS == OSWindows {
				out = append(out, g)
				break
			}
		}
	}
	return out
}
