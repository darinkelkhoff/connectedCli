package cli

import "testing"

func TestParseNoteLinks(t *testing.T) {
	htmlSrc := `<p>intro</p>
<h4>Sponsored by:</h4>
<ul>
<li><a href="https://sentry.io/">Sentry</a>: monitoring.</li>
<li><a href='https://steamclock.com/connected'>Steamclock &amp; co</a></li>
</ul>
<a href="https://relay.fm/connected/join"><h5>Get Connected Pro</h5></a>`
	links := parseNoteLinks(htmlSrc)
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %+v", len(links), links)
	}
	if links[0].Text != "Sentry" || links[0].URL != "https://sentry.io/" {
		t.Errorf("link 0 wrong: %+v", links[0])
	}
	// Entities decoded, inner tags stripped.
	if links[1].Text != "Steamclock & co" {
		t.Errorf("link 1 text wrong: %q", links[1].Text)
	}
	if links[2].Text != "Get Connected Pro" {
		t.Errorf("link 2 text (inner tag stripped) wrong: %q", links[2].Text)
	}
}
