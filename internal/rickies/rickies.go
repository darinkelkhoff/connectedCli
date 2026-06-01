package rickies

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const BaseURL = "https://rickies.net/api/v1"

// BillURL redirects to the current Bill of Rickies (on rickies.co).
const BillURL = "https://rickies.co/billof"

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
	Score      float64 `json:"score"`
	Conditions int    `json:"pick-conditions"`
}

// Hit reports whether the pick scored any points.
func (p Pick) Hit() bool { return p.Score > 0 }

// HostScore is a host's tally in a sub-game. Main games use Score; the flexies
// use Correct/Total.
type HostScore struct {
	Host    string `json:"host"`
	Score   float64 `json:"score,omitempty"`
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

// BillEntry is one line of the Bill of Rickies — a section heading or a rule.
type BillEntry struct {
	Heading bool   `json:"heading"`
	Text    string `json:"text"`
}

// The bill page (a date-versioned document) marks section headings with
// class="rule_type" and each rule as <div class="rule" data-start-date data-end-date><p>…</p>.
// Only rules whose [start,end] window contains "now" are currently in force.
var (
	billHeadingRe = regexp.MustCompile(`(?s)<h[1-4][^>]*class="rule_type[^"]*"[^>]*>(.*?)</h[1-4]>`)
	billRuleRe    = regexp.MustCompile(`(?s)data-start-date="(\d+)"[^>]*data-end-date="(\d+)"[^>]*>\s*<p>(.*?)</p>`)
	billTagRe     = regexp.MustCompile(`<[^>]*>`)
	billWsRe      = regexp.MustCompile(`\s+`)
	// Inline rule edits live in date-tagged <span>s; only those in force at
	// "now" should survive (expired/removed wording has a past end-date).
	billDatedSpanRe = regexp.MustCompile(`(?s)<span[^>]*data-start-date="(\d+)"[^>]*data-end-date="(\d+)"[^>]*>(.*?)</span>`)
)

func cleanBillText(s string) string {
	s = billTagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(billWsRe.ReplaceAllString(s, " "))
}

// resolveRuleText drops date-tagged spans that aren't in force at `now` (the
// page's inline edit history), then cleans the remaining text.
func resolveRuleText(s string, now int64) string {
	s = billDatedSpanRe.ReplaceAllStringFunc(s, func(m string) string {
		g := billDatedSpanRe.FindStringSubmatch(m)
		start, _ := strconv.ParseInt(g[1], 10, 64)
		end, _ := strconv.ParseInt(g[2], 10, 64)
		if now < start || now > end {
			return ""
		}
		return g[3]
	})
	return cleanBillText(s)
}

// ParseBill extracts the current Bill of Rickies (headings + rules in force at
// `now`, a Unix timestamp) from the page HTML, in document order.
func ParseBill(htmlStr string, now int64) []BillEntry {
	type located struct {
		pos   int
		entry BillEntry
	}
	var items []located
	for _, m := range billHeadingRe.FindAllStringSubmatchIndex(htmlStr, -1) {
		if text := cleanBillText(htmlStr[m[2]:m[3]]); text != "" {
			items = append(items, located{m[0], BillEntry{Heading: true, Text: text}})
		}
	}
	for _, m := range billRuleRe.FindAllStringSubmatchIndex(htmlStr, -1) {
		start, _ := strconv.ParseInt(htmlStr[m[2]:m[3]], 10, 64)
		end, _ := strconv.ParseInt(htmlStr[m[4]:m[5]], 10, 64)
		if now < start || now > end {
			continue
		}
		if text := resolveRuleText(htmlStr[m[6]:m[7]], now); text != "" {
			items = append(items, located{m[0], BillEntry{Heading: false, Text: text}})
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].pos < items[j].pos })
	out := make([]BillEntry, len(items))
	for i, it := range items {
		out[i] = it.entry
	}
	return out
}

// DiffBills compares two bill editions (rules only), returning the rules added
// in `to` and removed from `from`, each in their source order.
func DiffBills(from, to []BillEntry) (added, removed []BillEntry) {
	inFrom := map[string]bool{}
	for _, e := range from {
		if !e.Heading {
			inFrom[e.Text] = true
		}
	}
	inTo := map[string]bool{}
	for _, e := range to {
		if !e.Heading {
			inTo[e.Text] = true
		}
	}
	for _, e := range to {
		if !e.Heading && !inFrom[e.Text] {
			added = append(added, e)
		}
	}
	for _, e := range from {
		if !e.Heading && !inTo[e.Text] {
			removed = append(removed, e)
		}
	}
	return added, removed
}

// FirstParagraph returns the first rule (non-heading) entry's text, for a banner.
func FirstParagraph(entries []BillEntry) string {
	for _, e := range entries {
		if !e.Heading {
			return e.Text
		}
	}
	return ""
}

// FetchBillHTML downloads the raw Bill of Rickies page.
func FetchBillHTML(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, BillURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("bill: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bill: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FetchBill downloads and parses the Bill of Rickies in force at `now`.
func FetchBill(ctx context.Context, now int64) ([]BillEntry, error) {
	htmlStr, err := FetchBillHTML(ctx)
	if err != nil {
		return nil, err
	}
	return ParseBill(htmlStr, now), nil
}

// BillVersion is one historical edition of the bill, as listed by the site's
// history slider.
type BillVersion struct {
	Index    int    `json:"index"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Date     string `json:"date"`     // display form, e.g. "January 4, 2017"
	Datetime string `json:"datetime"` // ISO form, e.g. "2017-01-04"
	Value    int64  `json:"-"`        // the slider's filter timestamp (rickies_event_values)
}

// Unix returns the timestamp ParseBill should filter on for this edition — the
// site's own slider value when available, else noon on the version's date.
func (v BillVersion) Unix() int64 {
	if v.Value > 0 {
		return v.Value
	}
	t, err := time.Parse("2006-01-02", v.Datetime)
	if err != nil {
		return 0
	}
	return t.Add(12 * time.Hour).Unix()
}

var (
	billNamesArrRe  = regexp.MustCompile(`(?s)rickies_event_names\s*=\s*\[(.*?)\]`)
	billDatesArrRe  = regexp.MustCompile(`(?s)rickies_event_dates\s*=\s*\[(.*?)\]`)
	billUrlsArrRe   = regexp.MustCompile(`(?s)rickies_event_urls\s*=\s*\[(.*?)\]`)
	billValuesArrRe = regexp.MustCompile(`(?s)rickies_event_values\s*=\s*\[(.*?)\]`)
	jsStringRe      = regexp.MustCompile(`'((?:\\.|[^'])*)'`)
	billDatetimeRe  = regexp.MustCompile(`datetime="([^"]+)"`)
	billTimeTextRe  = regexp.MustCompile(`(?s)<time[^>]*>(.*?)</time>`)
)

// jsTimestampArray reads an array of (possibly quoted) integer timestamps.
func jsTimestampArray(htmlStr string, re *regexp.Regexp) []int64 {
	var out []int64
	for _, s := range jsStringArray(htmlStr, re) {
		if n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func jsStringArray(htmlStr string, re *regexp.Regexp) []string {
	m := re.FindStringSubmatch(htmlStr)
	if m == nil {
		return nil
	}
	var out []string
	for _, s := range jsStringRe.FindAllStringSubmatch(m[1], -1) {
		out = append(out, s[1])
	}
	return out
}

// ParseBillVersions extracts the bill's version history from the page's slider data.
func ParseBillVersions(htmlStr string) []BillVersion {
	names := jsStringArray(htmlStr, billNamesArrRe)
	dates := jsStringArray(htmlStr, billDatesArrRe)
	urls := jsStringArray(htmlStr, billUrlsArrRe)
	values := jsTimestampArray(htmlStr, billValuesArrRe)
	n := len(urls)
	if len(names) < n {
		n = len(names)
	}
	if len(dates) < n {
		n = len(dates)
	}
	versions := make([]BillVersion, 0, n)
	for i := 0; i < n; i++ {
		v := BillVersion{Index: i, Slug: urls[i], Name: cleanBillText(names[i])}
		if i < len(values) {
			v.Value = values[i]
		}
		if dm := billDatetimeRe.FindStringSubmatch(dates[i]); dm != nil {
			v.Datetime = dm[1]
		}
		if tm := billTimeTextRe.FindStringSubmatch(dates[i]); tm != nil {
			v.Date = cleanBillText(tm[1])
		}
		versions = append(versions, v)
	}
	return versions
}

// MatchBillVersion resolves a query (slug, index, or unique substring) to a version.
func MatchBillVersion(versions []BillVersion, query string) (BillVersion, bool) {
	if n, err := strconv.Atoi(query); err == nil {
		for _, v := range versions {
			if v.Index == n {
				return v, true
			}
		}
		return BillVersion{}, false
	}
	for _, v := range versions {
		if strings.EqualFold(v.Slug, query) {
			return v, true
		}
	}
	q := strings.ToLower(query)
	var hits []BillVersion
	for _, v := range versions {
		if strings.Contains(strings.ToLower(v.Slug), q) || strings.Contains(strings.ToLower(v.Name), q) {
			hits = append(hits, v)
		}
	}
	if len(hits) == 1 {
		return hits[0], true
	}
	return BillVersion{}, false
}
