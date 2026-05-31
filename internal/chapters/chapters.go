package chapters

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// headBytes is the size of the ranged GET (covers the ID3 tag on Connected).
const headBytes = 524288

// Fetch downloads the leading bytes of an MP3 and parses its chapters.
func Fetch(ctx context.Context, mp3URL string) ([]Chapter, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, mp3URL, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", headBytes-1))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch chapters: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("fetch chapters: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, headBytes))
	if err != nil {
		return nil, err
	}
	return ParseID3Chapters(data)
}

func First(chs []Chapter) Chapter { return chs[0] }
func Last(chs []Chapter) Chapter  { return chs[len(chs)-1] }

// At returns the chapter at index i (0-based).
func At(chs []Chapter, i int) (Chapter, bool) {
	if i < 0 || i >= len(chs) {
		return Chapter{}, false
	}
	return chs[i], true
}
