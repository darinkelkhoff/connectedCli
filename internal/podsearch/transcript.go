package podsearch

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Segment is one timestamped transcript line.
type Segment struct {
	Seconds int    `json:"seconds"`
	Text    string `json:"text"`
}

// Each transcript line renders as:
//
//	<p id="24" >  00:00:24  <a ...>◼</a> <a ...>►</a> It's a pleasure to be here...</p>
//
// The numeric id attribute is the second offset. pBlockRe captures id + inner HTML.
var pBlockRe = regexp.MustCompile(`(?s)<p id="(\d+)"[^>]*>(.*?)</p>`)

// playGlyphRe strips the ◼/► play-control glyph entities and stray timestamps.
var playGlyphRe = regexp.MustCompile(`&#9724;|&#9658;|◼|►|\d{2}:\d{2}:\d{2}`)

// parseTranscript extracts ordered timestamped segments from an episode page.
func parseTranscript(htmlStr string) []Segment {
	var segs []Segment
	for _, m := range pBlockRe.FindAllStringSubmatch(htmlStr, -1) {
		secs, _ := strconv.Atoi(m[1])
		inner := tagRe.ReplaceAllString(m[2], " ")     // drop <a> control tags
		inner = playGlyphRe.ReplaceAllString(inner, " ") // drop play glyphs + timestamp
		inner = html.UnescapeString(inner)               // &#39; -> '
		inner = strings.TrimSpace(wsRe.ReplaceAllString(inner, " "))
		if inner != "" {
			segs = append(segs, Segment{Seconds: secs, Text: inner})
		}
	}
	return segs
}

// TextInRange joins all segment text whose timestamp falls in [startSec, endSec].
func TextInRange(segs []Segment, startSec, endSec int) string {
	var parts []string
	for _, s := range segs {
		if s.Seconds >= startSec && s.Seconds <= endSec {
			parts = append(parts, s.Text)
		}
	}
	return strings.Join(parts, " ")
}

// listingRe matches "<a href="/episodes/7950">601:" — internal id then episode number.
var listingRe = regexp.MustCompile(`/episodes/(\d+)"[^>]*>\s*(\d+):`)

// parseShowListing maps Connected episode number -> David Smith internal id.
func parseShowListing(htmlStr string) map[int]int {
	m := map[int]int{}
	for _, mm := range listingRe.FindAllStringSubmatch(htmlStr, -1) {
		id, _ := strconv.Atoi(mm[1])
		num, _ := strconv.Atoi(mm[2])
		if num > 0 {
			m[num] = id
		}
	}
	return m
}

// firstListedEpisode returns the episode number of the first listing entry.
// The listing is newest-first, so this is the newest transcribed episode.
// (Using document order avoids treating a year-prefixed title like
// "2016: Big, Heavy and Vague" as the newest episode number.)
func firstListedEpisode(htmlStr string) int {
	mm := listingRe.FindStringSubmatch(htmlStr)
	if mm == nil {
		return 0
	}
	num, _ := strconv.Atoi(mm[2])
	return num
}

// FetchTranscript downloads and parses an episode transcript by internal id.
func FetchTranscript(ctx context.Context, internalID int) ([]Segment, error) {
	u := fmt.Sprintf("%s/episodes/%d", BaseURL, internalID)
	body, err := getBody(ctx, u)
	if err != nil {
		return nil, err
	}
	return parseTranscript(string(body)), nil
}

// InternalID resolves a Connected episode number to a David Smith internal id.
func InternalID(ctx context.Context, episodeNumber int) (int, error) {
	u := fmt.Sprintf("%s/shows/%d", BaseURL, ConnectedShow)
	body, err := getBody(ctx, u)
	if err != nil {
		return 0, err
	}
	m := parseShowListing(string(body))
	id, ok := m[episodeNumber]
	if !ok {
		newest := firstListedEpisode(string(body))
		if newest > 0 {
			return 0, fmt.Errorf("no transcript yet for episode %d (newest transcribed is %d; transcripts lag the feed). Try --play for audio, or --episode %d", episodeNumber, newest, newest)
		}
		return 0, fmt.Errorf("no transcript found for episode %d", episodeNumber)
	}
	return id, nil
}

// NewestTranscribed returns the highest episode number David Smith has transcribed.
func NewestTranscribed(ctx context.Context) (int, error) {
	body, err := getBody(ctx, fmt.Sprintf("%s/shows/%d", BaseURL, ConnectedShow))
	if err != nil {
		return 0, err
	}
	newest := firstListedEpisode(string(body))
	if newest == 0 {
		return 0, fmt.Errorf("could not determine newest transcribed episode")
	}
	return newest, nil
}

func getBody(ctx context.Context, u string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", u, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
