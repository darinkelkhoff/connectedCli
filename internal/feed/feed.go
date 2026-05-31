package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const FeedURL = "https://relay.fm/connected/feed"

// Episode is one Connected episode.
type Episode struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	MP3URL    string    `json:"mp3Url"`
	Link      string    `json:"link"`
	Summary   string    `json:"summary,omitempty"`
	NotesHTML string    `json:"-"` // raw content:encoded; parsed by the notes command
}

type rss struct {
	Channel struct {
		Items []struct {
			Title     string `xml:"title"`
			Link      string `xml:"link"`
			PubDate   string `xml:"pubDate"`
			Episode   int    `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd episode"`
			Summary   string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
			Notes     string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
			Enclosure struct {
				URL string `xml:"url,attr"`
			} `xml:"enclosure"`
		} `xml:"item"`
	} `xml:"channel"`
}

// Parse turns raw RSS bytes into episodes.
func Parse(data []byte) ([]Episode, error) {
	var doc rss
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}
	out := make([]Episode, 0, len(doc.Channel.Items))
	for _, it := range doc.Channel.Items {
		date, _ := time.Parse(time.RFC1123Z, it.PubDate)
		out = append(out, Episode{
			Number:    it.Episode,
			Title:     it.Title,
			Date:      date,
			MP3URL:    it.Enclosure.URL,
			Link:      it.Link,
			Summary:   strings.TrimSpace(it.Summary),
			NotesHTML: it.Notes,
		})
	}
	return out, nil
}

// Fetch downloads and parses the live feed.
func Fetch(ctx context.Context) ([]Episode, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, FeedURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch feed: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Latest returns the highest-numbered episode.
func Latest(eps []Episode) Episode {
	var best Episode
	for _, e := range eps {
		if e.Number >= best.Number {
			best = e
		}
	}
	return best
}

// ByNumber finds an episode by its Connected number.
func ByNumber(eps []Episode, n int) (Episode, bool) {
	for _, e := range eps {
		if e.Number == n {
			return e, true
		}
	}
	return Episode{}, false
}
