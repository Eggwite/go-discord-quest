package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const discordURL = "https://discord.com/api/applications/detectable"

func FetchDetectable(ctx context.Context) ([]Game, error) {
	games, _, err := FetchDetectableWithTrace(ctx)
	return games, err
}

func FetchDetectableWithTrace(ctx context.Context) ([]Game, []string, error) {
	trace := make([]string, 0, 2)

	trace = append(trace, "* Fetching from Discord API...")
	games, err := fetchURL(ctx, discordURL)
	if err != nil {
		trace = append(trace, fmt.Sprintf("  - Failed: %v", err))
		return nil, trace, errors.New("could not fetch game list from Discord API")
	}

	trace = append(trace, "  + Done")
	return filterWindows(games), trace, nil
}

func fetchURL(ctx context.Context, url string) ([]Game, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
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

	var games []Game
	if err := json.Unmarshal(body, &games); err != nil {
		return nil, err
	}
	if len(games) == 0 {
		return nil, errors.New("empty game list")
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
