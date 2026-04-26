package search

import (
	"sort"
	"strings"

	"github.com/Eggwite/go-discord-quest/internal/discord"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func Search(games []discord.Game, query string) []discord.Game {
	if strings.TrimSpace(query) == "" {
		return games
	}

	q := stripSymbols(strings.ToLower(query))
	type scoredGame struct {
		game  discord.Game
		score float64
	}

	scored := make([]scoredGame, 0, len(games))
	for _, g := range games {
		nameScore := rankScore(q, stripSymbols(strings.ToLower(g.Name)))
		aliasScore := bestAliasScore(q, g.Aliases)
		exeScore := bestExeScore(q, g.Executables)
		score := 0.7*nameScore + 0.2*aliasScore + 0.1*exeScore
		if score > 0 {
			scored = append(scored, scoredGame{game: g, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].game.Name < scored[j].game.Name
		}
		return scored[i].score > scored[j].score
	})

	out := make([]discord.Game, 0, len(scored))
	for _, s := range scored {
		out = append(out, s.game)
	}
	return out
}

func bestAliasScore(query string, aliases []string) float64 {
	best := 0.0
	for _, alias := range aliases {
		s := rankScore(query, stripSymbols(strings.ToLower(alias)))
		if s > best {
			best = s
		}
	}
	return best
}

func bestExeScore(query string, executables []discord.GameExecutable) float64 {
	best := 0.0
	for _, exe := range executables {
		s := rankScore(query, stripSymbols(strings.ToLower(exe.Name)))
		if s > best {
			best = s
		}
	}
	return best
}

func rankScore(query, candidate string) float64 {
	distance := fuzzy.RankMatchNormalized(query, candidate)
	if distance < 0 {
		return 0
	}
	return 1.0 / float64(distance+1)
}

func stripSymbols(s string) string {
	replacer := strings.NewReplacer("™", "", "©", "", "®", "")
	return replacer.Replace(s)
}
