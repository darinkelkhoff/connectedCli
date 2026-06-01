package rickies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const BaseURL = "https://rickies.net/api/v1"

// Award is a won title (chairman, flexie, etc.).
type Award struct {
	GameName string `json:"game-name"`
	Date     string `json:"date"`
	Winner   string `json:"winner"`
	Title    string `json:"title"`
	Social   string `json:"social"`
}

// Winners is the current standings payload (winners.json).
type Winners struct {
	Titles        map[string][]string `json:"titles"`
	Annual        Award               `json:"annual"`
	Keynote       Award               `json:"keynote"`
	Flexies       Award               `json:"flexies"`
	AnnualFlexies Award               `json:"annual-flexies"`
}

// EpisodeGames maps a Connected episode to its relevant Rickies games.
type EpisodeGames struct {
	Episode       int      `json:"episode"`
	Permalink     string   `json:"permalink"`
	RelevantGames []string `json:"relevant-games"`
}

func ParseWinners(data []byte) (Winners, error) {
	var w Winners
	if err := json.Unmarshal(data, &w); err != nil {
		return Winners{}, fmt.Errorf("parse winners: %w", err)
	}
	return w, nil
}

func ParseEpisodes(data []byte) ([]EpisodeGames, error) {
	var eps []EpisodeGames
	if err := json.Unmarshal(data, &eps); err != nil {
		return nil, fmt.Errorf("parse episodes: %w", err)
	}
	return eps, nil
}

func getJSON(ctx context.Context, path string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, BaseURL+path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rickies: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rickies: status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// FetchWinners gets the current standings.
func FetchWinners(ctx context.Context) (Winners, error) {
	data, err := getJSON(ctx, "/winners.json")
	if err != nil {
		return Winners{}, err
	}
	return ParseWinners(data)
}

// FetchEpisodes gets the episode↔games map.
func FetchEpisodes(ctx context.Context) ([]EpisodeGames, error) {
	data, err := getJSON(ctx, "/episodes.json")
	if err != nil {
		return nil, err
	}
	return ParseEpisodes(data)
}

// GameRef is an index entry that points to a game (not a pick/topic/coinflip).
type GameRef struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Pick is one prediction made by a host.
type Pick struct {
	Host       string `json:"host"`
	Text       string `json:"text"`
	Notes      string `json:"notes,omitempty"`
	Type       string `json:"type,omitempty"`
	Score      int    `json:"score"`
	Conditions int    `json:"pick-conditions"`
}

// Hit reports whether the pick scored any points.
func (p Pick) Hit() bool { return p.Score > 0 }

// HostScore is a host's tally in a sub-game. Main games use Score; the flexies
// use Correct/Total.
type HostScore struct {
	Host    string `json:"host"`
	Score   int    `json:"score,omitempty"`
	Correct int    `json:"correct,omitempty"`
	Total   int    `json:"total,omitempty"`
}

// SubGame is one round of a game (the main game, or the flexies).
type SubGame struct {
	Picks       []Pick      `json:"picks"`
	Scores      []HostScore `json:"scores"`
	Winner      string      `json:"winner"`
	CharityName string      `json:"charity-name,omitempty"`
	CharityURL  string      `json:"charity-url,omitempty"`
	Donation    int         `json:"donation,omitempty"`
	Donator     string      `json:"donator,omitempty"`
}

// Game is the full detail of a single Rickies game.
type Game struct {
	Name       string  `json:"name"`
	GameType   string  `json:"game-type"`
	DatePicked string  `json:"date-picked"`
	DateGraded string  `json:"date-graded"`
	MainGame   SubGame `json:"main-game"`
	Flexies    SubGame `json:"the-flexies"`
}

// ParseGameIndex returns only the game entries from index.json. The index also
// contains picks, topics, and coin flips; those are filtered out by path.
func ParseGameIndex(data []byte) ([]GameRef, error) {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	var games []GameRef
	for name, path := range m {
		if strings.HasPrefix(path, "/game/") {
			games = append(games, GameRef{Name: name, Path: path})
		}
	}
	return games, nil
}

func ParseGame(data []byte) (Game, error) {
	var g Game
	if err := json.Unmarshal(data, &g); err != nil {
		return Game{}, fmt.Errorf("parse game: %w", err)
	}
	return g, nil
}

// FetchGameIndex returns the list of Rickies games.
func FetchGameIndex(ctx context.Context) ([]GameRef, error) {
	data, err := getJSON(ctx, "/index.json")
	if err != nil {
		return nil, err
	}
	return ParseGameIndex(data)
}

// FetchGame fetches a game's detail by its index path (e.g. "/game/39.json").
func FetchGame(ctx context.Context, path string) (Game, error) {
	data, err := getJSON(ctx, path)
	if err != nil {
		return Game{}, err
	}
	return ParseGame(data)
}
