package search

import (
	"sort"
	"strings"
	"time"

	"github.com/Eggwite/go-discord-quest/internal/discord"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// SearchableGame stores pre-processed strings for zero-allocation comparisons
type SearchableGame struct {
	Game         discord.Game
	SearchTarget string
}

// PrepareSearchData converts raw Discord games into search-optimised structures
// Performs lowercase transformations once during initialisation
func PrepareSearchData(games []discord.Game) []SearchableGame {
	out := make([]SearchableGame, len(games))
	for i, g := range games {
		var sb strings.Builder
		sb.WriteString(strings.ToLower(g.Name))
		for _, a := range g.Aliases {
			sb.WriteString(" " + strings.ToLower(a))
		}
		out[i] = SearchableGame{
			Game:         g,
			SearchTarget: sb.String(),
		}
	}
	return out
}

// Search filters the dataset and returns results along with execution duration
// Utilises a fast-path substring check before applying fuzzy ranking
func Search(data []SearchableGame, query string) ([]discord.Game, time.Duration) {
	start := time.Now()

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		out := make([]discord.Game, len(data))
		for i, d := range data {
			out[i] = d.Game
		}
		return out, time.Since(start)
	}

	q := strings.ToLower(trimmed)
	type scoredGame struct {
		game  discord.Game
		score float64
	}

	scored := make([]scoredGame, 0, 512)
	for _, item := range data {
		// Fast-path check to avoid fuzzy overhead on non-matching entries
		if !strings.Contains(item.SearchTarget, q) {
			continue
		}

		distance := fuzzy.RankMatchNormalized(q, item.SearchTarget)
		if distance != -1 {
			score := 1.0 / float64(distance+1)
			scored = append(scored, scoredGame{game: item.Game, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].game.Name < scored[j].game.Name
		}
		return scored[i].score > scored[j].score
	})

	out := make([]discord.Game, len(scored))
	for i, s := range scored {
		out[i] = s.game
	}

	return out, time.Since(start)
}
