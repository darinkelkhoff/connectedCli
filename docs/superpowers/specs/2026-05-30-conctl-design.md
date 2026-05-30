# conctl — the Connected CLI

**Date:** 2026-05-30
**Status:** Design approved, pending spec review

## Spirit

`conctl` is a gag-but-genuinely-functional command-line tool for the
[Connected](https://relay.fm/connected) podcast (Relay FM — Myke Hurley, Stephen
Hackett, Federico Viticci). It was prompted by Stephen riffing "connected cli…
connected --exit" on episode 605 (the episode that discussed
[remctl](https://www.macstories.net/stories/introducing-remctl-the-power-user-reminders-cli-for-macos-and-ai-agents/),
the power-user-meets-AI-agent Reminders CLI).

Like remctl, conctl is built to serve **both** a human power user and an AI agent:
pretty human output by default, `--json` everywhere for agents. The comedy is the
frame; the core actually works.

Scope for v1: **tight gag, real core.** The comedic `--intro`/`--exit` framing
sits on top of a real chapter engine, plus two standalone real features
(transcript search, the Rickies).

## Goals

- A single, giftable macOS binary installable with one Homebrew command, hosted
  entirely on GitHub.
- The `--intro`/`--exit` bit lands because it is backed by real episode data.
- Transcript search and Rickies standings work against live endpoints.
- `--json` on every command for agent use.
- Local-first AI: an optional `--llm` mode that reuses an AI CLI the user already
  has installed (no API keys, no new services).

## Non-goals (YAGNI)

- No "next intro" prediction feature (explicitly cut).
- No login/membership (Connected Pro) features.
- No write operations of any kind — conctl is read-only.
- No hosting anything outside GitHub (no PyPI, no npm registry, no web service).
- No bundled/managed local model — conctl never installs an LLM; it only uses one
  that is already present.

## Distribution

- **Language:** Go. Single static binary. `cobra` for the command tree.
- **Install:** `brew install dkelkhoff/tap/conctl`. A sibling GitHub repo
  `dkelkhoff/homebrew-tap` holds the formula; the binary is attached to GitHub
  Releases (built/published with GoReleaser). 100% GitHub — no other registry.
- **Identifier:** `com.kelkhoff.conctl` where a bundle/package id is needed.
- **Runtime dependencies:**
  - Core (search, rickies, chapters list, intro/exit text + `--say` + `--json`):
    **none** beyond macOS (`say` is built in).
  - `--play` (real chapter audio): requires `ffmpeg`/`ffplay`. Declared as
    `depends_on "ffmpeg"` in the Homebrew formula; a friendly error if invoked
    when missing.
  - `--llm`: requires a detected local AI CLI (see LLM Providers). Friendly error
    if none found.

## Verified data sources

All endpoints below were verified live during design (2026-05-30).

### 1. Transcript search — David Smith's Podcast Search

- Connected is **show 17**: `https://podcastsearch.david-smith.org/shows/17`
- Search is a per-show GET form:
  `GET https://podcastsearch.david-smith.org/shows/17?term=<query>`
- Results are HTML. Each hit is an anchor of the form
  `/episodes/<internalId>#<seconds>` where `<seconds>` is the **exact second
  offset** into the episode (verified: timestamp `01:40:25` ↔ anchor `#6025`).
  The matched phrase is wrapped in `<b>…</b>` within the snippet line, with the
  `HH:MM:SS` timestamp adjacent.
- Episode transcript pages (`/episodes/<internalId>`) render timestamped
  segment lines (`HH:MM:SS` + play button + text). Used for slicing chapter text
  by time range.
- `<internalId>` is David Smith's internal id, **not** the Connected episode
  number. Mapping (Connected ep # ↔ internalId) is done by scraping the
  `/shows/17` listing / episode-page titles, which contain `"605: Being
  Completely Awesome"`-style titles.

### 2. The Rickies — rickies.co API

- `GET https://rickies.co/api/chairmen.json` → JSON:
  ```json
  {
    "keynote_chairman": { "name": "...", "last_name": "...", "location": "...", "twitter": "...", "memoji": "..." },
    "annual_chairman":  { "name": "...", "last_name": "...", "location": "...", "twitter": "...", "memoji": "..." }
  }
  ```
- Open-source project: https://github.com/lexpostma/Rickies. Leaderboard data is
  **not** in the API today (site-only); v1 ships the chairmen endpoint and notes
  leaderboard as a possible future scrape.

### 3. Episodes & chapters — Connected RSS + MP3 ID3

- Feed: `https://www.relay.fm/connected/feed` (302 → `https://relay.fm/connected/feed`).
  - `<item>` has `<itunes:episode>` (e.g. `605`), `<title>` (`"605: Being
    Completely Awesome"`), pubDate, and an `<enclosure url="…mp3" .../>`
    (podtrac-prefixed Libsyn MP3).
  - The feed does **not** contain `<podcast:chapters>` or PSC tags.
- **Chapters live in the MP3 as ID3v2 chapter frames** (`CTOC` + per-chapter
  `CHAP`, each carrying a `TIT2` title). Verified on ep 605: 1 `CTOC`, 9 `CHAP`,
  10 `TIT2`, all within the first 400 KB.
  - conctl reads chapters via a **ranged GET** of the start of the MP3 (HTTP 206;
    fetch the ID3v2 tag region — size is declared in the ID3 header — ~first
    256–512 KB is sufficient) and parses `CHAP` frames into
    `{index, title, startMs, endMs}`. No full download.

## Command surface

Global flags: `--json` (structured output for agents), `--version`, `--help`
(help text styled as a show rundown: Follow-up, Topics, The Rickies…).

Episode targeting (shared): default target is the **latest** episode from the
feed; override with `--episode <n>` or a positional ep number where natural.

### Chapter engine — the unifying model

`--intro` = the **first** chapter; `--exit` = the **last** chapter. Both are
presets over one chapter engine and the shared render modes below.

- `conctl chapters [ep]` — list chapters (index, title, start/end) for an episode.
- `conctl play [ep] [--chapter N | --first | --last]` — play a chapter's audio.
- `conctl --intro [ep] [render flags]` — first chapter.
- `conctl --exit  [ep] [render flags]` — last chapter. The marquee bit.

### Render modes (shared across intro / exit / play / chapters)

How a selected chapter is emitted:

- **(default) text** — the chapter's transcript text, sliced from the David Smith
  episode page by the chapter's `[start,end]` time range. Instant, zero deps,
  works over SSH and for agents. (`chapters` with no other flag lists titles+times.)
- `--play` — stream just that chapter's audio segment via
  `ffplay -ss <start> -t <dur> <mp3url>`. Requires ffmpeg.
- `--say` — pipe the chapter text to macOS `say`. Built in, zero deps.
  (`conctl --exit --say` reads the show's goodbye aloud.)
- `--json` — structured chapter + content.
- `--short` — skip the chapter entirely; print only the quick host sign-offs:
  `Arrivederci. Cheerio. Bye, y'all.` (exit) or a one-line cold-open (intro).
- `--llm` — generate content from the chapter/episode via a local AI CLI (below).
  Combinable with `--say` and `--json`. `--prompt <name>` selects a preset.

`--intro`/`--exit` **default to text**. `--play`, `--say`, `--llm`, `--short`,
`--json` are opt-in. Flags compose where sensible (e.g. `--llm --say`).

### Standalone real features

- `conctl search <query> [--limit N] [--json]` — Connected transcript search.
  Human output: ranked hits with episode, `HH:MM:SS`, bolded snippet, and a
  clickable deep-link to the exact second. `--json`: array of
  `{episodeInternalId, episodeNumber?, episodeTitle?, time, seconds, snippet, url}`.
- `conctl rickies [--json]` — current keynote + annual **Chairman** (name,
  location, memoji link) from `chairmen.json`. `--json` passes the upstream shape
  through (normalized).

## LLM providers (`--llm`)

Reuse an AI CLI the user already has — no API keys, no install by conctl.
`internal/llm` exposes one interface:

```go
type Provider interface {
    Name() string
    Available() bool                                   // exec.LookPath / probe
    Generate(ctx, system, user string) (string, error)
}
```

Detection priority (first available wins; `--provider <name>` to force):

1. **claude** (Claude Code CLI) — headless print mode: `claude -p "<prompt>"`.
2. **codex** (Codex CLI) — non-interactive: `codex exec "<prompt>"`.
3. **Apple Foundation Models** (on-device, macOS 26) — **flagged stretch**; needs
   a tiny bundled Swift helper, so it is the one provider that breaks pure-Go.
   Not in the core v1 unless promoted at review.
4. Optional fallbacks, never assumed: `llm` CLI, any OpenAI-compatible
   `localhost` endpoint. (Ollama is **not** privileged or assumed.)

If none are available: a friendly message naming what conctl looked for and how
to enable one. Exact invocation flags for `claude`/`codex` are verified at
implementation time; the interface isolates that detail.

### Prompt presets

A small curated set, selected with `--prompt <name>`; each is a
well-engineered system+user prompt fed the chapter (or whole-episode) transcript:

- `conclusion` *(default for `--exit --llm`)* — a warm, in-the-show's-voice
  closing summary of the episode.
- `cold-open` *(default for `--intro --llm`)* — a punchy generated intro.
- `recap` — concise bulleted recap of the whole episode.
- `style-federico` / `style-myke` / `style-stephen` — outro in that host's voice.
- `haiku` — the episode as a haiku.

Presets are data (embedded), easy to extend.

## Internal package layout

- `internal/feed` — fetch/parse Connected RSS → `[]Episode{number,title,date,mp3URL}`;
  `Latest()`, `ByNumber(n)`.
- `internal/chapters` — ranged MP3 GET + ID3v2 `CTOC`/`CHAP` parse →
  `[]Chapter{index,title,startMs,endMs}`.
- `internal/podsearch` — `Search(term)`; episode transcript fetch + time-range
  text slice; ep# ↔ internalId mapping.
- `internal/rickies` — `chairmen.json` client.
- `internal/audio` — `ffplay` segment wrapper (+ missing-ffmpeg detection) and
  `say` wrapper.
- `internal/llm` — provider interface + detectors + embedded prompt presets.
- `internal/render` — pretty vs JSON emitters; golden-file friendly.
- `cmd/conctl` — cobra command tree wiring the above.

Each integration is isolated behind a typed client and is unit-testable from
recorded fixtures.

## `--json` output (illustrative shapes)

- search hit: `{ "episodeInternalId": 7950, "episodeNumber": 605, "episodeTitle": "...", "time": "01:40:25", "seconds": 6025, "snippet": "...", "url": "https://podcastsearch.david-smith.org/episodes/7950#6025" }`
- chapter: `{ "index": 8, "title": "...", "start": "01:38:00", "startMs": 5880000, "endMs": 6300000 }`
- chapter content (text/llm): chapter object + `{ "text": "...", "source": "transcript|llm", "provider": "claude" }`
- rickies: `{ "keynoteChairman": {...}, "annualChairman": {...} }`

## Error handling

- Network failures: clear message + non-zero exit; `--json` emits
  `{ "error": "..." }`.
- `--play` without ffmpeg: explain and suggest `brew install ffmpeg` (or that the
  formula should have pulled it).
- `--llm` with no provider: list detectors tried and how to enable one.
- Episode/chapter not found: name the valid range (latest episode number).
- ID3 parse finding no chapters: fall back to whole-episode text/audio with a note.

## Testing

- Table tests against **saved fixtures** captured from the real endpoints:
  a `/shows/17?term=…` results page, an episode transcript page, `chairmen.json`,
  a feed `<item>`, and a truncated MP3 ID3 region with known `CHAP` frames.
- Golden-file tests for pretty (human) output.
- No live network in tests.
- LLM providers tested with a fake provider implementing the interface; detection
  logic tested by stubbing `exec.LookPath`.

## Flagged decisions for spec review

1. **Apple Foundation Models** as a provider is currently a *stretch* (it adds a
   Swift helper and breaks pure-Go). Promote to core?
2. **Default render** for `--intro`/`--exit` is **text** (approved). `--play`,
   `--say`, `--llm` are opt-in.
3. Rickies **leaderboard** scraping is out of v1 (API only exposes chairmen).

## Suggested build order

1. Skeleton: cobra tree, `--json` plumbing, `--version`, rundown help.
2. `internal/feed` + `conctl latest`/episode targeting.
3. `internal/podsearch` + `conctl search`.
4. `internal/rickies` + `conctl rickies`.
5. `internal/chapters` (ID3 parse) + `conctl chapters` + transcript text slice.
6. Render modes + `--intro`/`--exit` (text, `--short`, `--json`).
7. `internal/audio` + `--play` / `--say`.
8. `internal/llm` + `--llm` + prompt presets.
9. GoReleaser + Homebrew tap.
