package search

import (
	"sort"
	"strings"
	"time"

	"github.com/Eggwite/go-discord-quest/internal/discord"
)

// SearchableGame stores pre-processed strings for zero-allocation comparisons
type SearchableGame struct {
	Game         discord.Game
	SearchTarget string
}
type scoredGame struct {
	game  discord.Game
	score float64
}

type scoredGames []scoredGame

func (s scoredGames) Len() int      { return len(s) }
func (s scoredGames) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s scoredGames) Less(i, j int) bool {
	if s[i].score == s[j].score {
		return s[i].game.Name < s[j].game.Name
	}
	return s[i].score > s[j].score
}

// PrepareSearchData converts raw Discord games into search-optimised structures.
// Lowercase transformations are performed once at initialisation to avoid
// repeating them on every search call.
func PrepareSearchData(games []discord.Game) []SearchableGame {
	out := make([]SearchableGame, len(games))
	for i, g := range games {
		var sb strings.Builder
		name := strings.ToLower(g.Name)
		totalLen := len(name)
		for _, a := range g.Aliases {
			totalLen += 1 + len(a)
		}
		sb.Grow(totalLen)
		sb.WriteString(name)
		for _, a := range g.Aliases {
			sb.WriteByte(' ')
			sb.WriteString(strings.ToLower(a))
		}
		out[i] = SearchableGame{
			Game:         g,
			SearchTarget: sb.String(),
		}
	}
	return out
}

// score returns a relevance score in the range (0, 1] for a query against a
// search target, or 0 if the target does not contain the query at all.
//
// Scoring tiers (descending priority):
//  1. Exact full match                  → 1.0
//  2. Match at the start of the string  → 0.9 - positional penalty
//  3. Match elsewhere in the string     → 0.6 - positional penalty
//
// The positional penalty is the normalised index of the match, scaled to
// a small fraction of the tier range so that earlier matches rank higher
// within the same tier.
func score(q, target string) float64 {
	if target == q {
		return 1.0
	}

	idx := strings.Index(target, q)
	if idx == -1 {
		return 0
	}

	// Normalised position: 0.0 at start, approaching 1.0 at end
	pos := float64(idx) / float64(len(target))

	if idx == 0 {
		// Prefix match: scores between 0.8 and 0.9
		return 0.9 - pos*0.1
	}

	// Mid-string match: scores between 0.5 and 0.6
	return 0.6 - pos*0.1
}

// Search filters the dataset against the query and returns matching games
// ranked by relevance, along with the execution duration.
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

	scored := make([]scoredGame, 0, len(data))

	for _, item := range data {
		s := score(q, item.SearchTarget)
		if s > 0 {
			scored = append(scored, scoredGame{game: item.Game, score: s})
		}
	}

	sort.Sort(scoredGames(scored))

	out := make([]discord.Game, len(scored))
	for i, s := range scored {
		out[i] = s.game
	}

	return out, time.Since(start)
}
