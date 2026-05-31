package rickies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// FetchGames gets the raw index.json (game name -> details) as a generic map.
func FetchGames(ctx context.Context) (map[string]any, error) {
	data, err := getJSON(ctx, "/index.json")
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	return m, nil
}
