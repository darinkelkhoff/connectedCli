package podsearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	BaseURL       = "https://podcastsearch.david-smith.org"
	ConnectedShow = 17
)

// Hit is one transcript search match.
type Hit struct {
	EpisodeInternalID int    `json:"episodeInternalId"`
	EpisodeNumber     int    `json:"episodeNumber,omitempty"`
	EpisodeTitle      string `json:"episodeTitle,omitempty"`
	Seconds           int    `json:"seconds"`
	Time              string `json:"time"`
	Snippet           string `json:"snippet"`
	URL               string `json:"url"`
}

// A result row renders as:
//
//	<h2>601: I Love Wrists</h2>
//	<p>
//	    01:40:25
//	    <a href="/episodes/7950#6025"> ...snippet... </a>
//	</p>
//
// hitBlockRe captures timestamp, internal id, seconds, and snippet together.
var hitBlockRe = regexp.MustCompile(`(?s)(\d{2}:\d{2}:\d{2})\s*<a href="/episodes/(\d+)#(\d+)">(.*?)</a>`)

// h2Re captures the episode heading "<number>: <title>".
var h2Re = regexp.MustCompile(`<h2>\s*(\d+):\s*([^<]*?)\s*</h2>`)

var tagRe = regexp.MustCompile(`<[^>]+>`)
var wsRe = regexp.MustCompile(`\s+`)

func cleanText(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

// parseSearchResults extracts hits from a results HTML page, attributing each
// to the nearest preceding episode heading.
func parseSearchResults(html string) []Hit {
	type marker struct {
		pos   int
		num   int
		title string
	}
	var markers []marker
	for _, m := range h2Re.FindAllStringSubmatchIndex(html, -1) {
		num, _ := strconv.Atoi(html[m[2]:m[3]])
		markers = append(markers, marker{pos: m[0], num: num, title: html[m[4]:m[5]]})
	}
	epAt := func(pos int) (int, string) {
		num, title := 0, ""
		for _, mk := range markers {
			if mk.pos < pos {
				num, title = mk.num, mk.title
			} else {
				break
			}
		}
		return num, title
	}

	var hits []Hit
	for _, m := range hitBlockRe.FindAllStringSubmatchIndex(html, -1) {
		ts := html[m[2]:m[3]]
		id, _ := strconv.Atoi(html[m[4]:m[5]])
		secs, _ := strconv.Atoi(html[m[6]:m[7]])
		snippet := cleanText(html[m[8]:m[9]])
		num, title := epAt(m[0])
		hits = append(hits, Hit{
			EpisodeInternalID: id,
			EpisodeNumber:     num,
			EpisodeTitle:      title,
			Seconds:           secs,
			Time:              ts,
			Snippet:           snippet,
			URL:               fmt.Sprintf("%s/episodes/%d#%d", BaseURL, id, secs),
		})
	}
	return hits
}

// Search runs a transcript search against the Connected show.
func Search(ctx context.Context, term string, limit int) ([]Hit, error) {
	u := fmt.Sprintf("%s/shows/%d?term=%s", BaseURL, ConnectedShow, url.QueryEscape(term))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	hits := parseSearchResults(string(body))
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}
