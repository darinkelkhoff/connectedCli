---
name: conctl
description: Use when the user asks about the Connected podcast (Relay FM — Myke Hurley, Federico Viticci, Stephen Hackett) — searching what was said and when, listing or playing an episode's chapters, getting show notes/links, checking The Rickies (current Chairmen, titles, games), or fetching an episode's intro/closing. conctl is a read-only CLI; every command supports --json.
---

# conctl — the Connected CLI (agent guide)

`conctl` is a read-only macOS CLI for the [Connected](https://relay.fm/connected)
podcast. It reads only public data (the RSS feed, David Smith's transcript site,
the rickies.net API) and never writes anything, so it is safe to run freely.

**Always pass `--json`.** Human output is for terminals; `--json` is stable,
goes to stdout, and is what you should parse. Notes/spinners/errors go to stderr.

## Check availability

```bash
command -v conctl || echo "not installed: brew install dkelkhoff/tap/conctl"
conctl --version
```

## Episode targeting

Most commands take an optional episode as a positional number. With none, they
use the **latest** episode:

```bash
conctl chapters 605
conctl intro 601
```

## Commands and JSON shapes

### Search transcripts

```bash
conctl search "vision pro" --limit 10 --json
```
Returns an array of hits:
```json
[{ "episodeInternalId": 7950, "episodeNumber": 601, "episodeTitle": "I Love Wrists",
   "time": "01:40:25", "seconds": 6025, "snippet": "…",
   "url": "https://podcastsearch.david-smith.org/episodes/7950#6025" }]
```
`seconds` is the exact offset; `url` deep-links to that moment.

### Latest episode

```bash
conctl latest --json
```
```json
{ "number": 605, "title": "Being Completely Awesome", "date": "…",
  "mp3Url": "…", "link": "https://relay.fm/connected/605", "summary": "…" }
```
Add `--full` to also include the episode's chapters and show-note links:
`conctl latest --full --json` adds `chapters[]` and `links[]`.

### Chapters

```bash
conctl chapters 605 --json
```
```json
[{ "index": 7, "title": "RemCTL", "startMs": 3009542, "endMs": 4318371 }]
```

### Intro / closing chapter

`conctl intro [ep]` = first chapter, `conctl exit [ep]` = last chapter. (The
`--intro`/`--exit` flag form on the bare `conctl` command also works — that's the
show's running joke — but prefer the subcommands.) Default output is the
chapter's transcript text; add `--json`:

```bash
conctl intro 601 --json
conctl exit 601 --json
```
```json
{ "episode": 601, "chapter": { "index": 0, "title": "…", "startMs": 0, "endMs": 73556 }, "text": "…" }
```
Modes (choose at most one): `--play` (audio, needs ffmpeg), `--say` (macOS TTS),
`--llm` (rewrite via a local AI CLI), `--short` (just the sign-offs). `--prompt`
selects an `--llm` preset: `conclusion`, `cold-open`, `recap`, `style-federico`,
`style-myke`, `style-stephen`, `haiku`.

### Show notes

```bash
conctl notes 605 --json
```
```json
{ "episode": 605, "title": "…", "summary": "…",
  "links": [{ "text": "Sentry", "url": "https://sentry.io/" }] }
```

### The Rickies

```bash
conctl rickies --json          # current standings
conctl rickies titles --json   # per-host titles
conctl rickies games --json    # [{ "name": "2026 March Rickies", "path": "/game/39.json" }, …]
conctl rickies game "2026 March" --json   # one game in full detail
conctl rickies bill --json                # the current Bill of Rickies (rules)
```
`rickies bill --json` returns the current rules in order as
`[{ "heading": bool, "text": "…" }]` (headings mark sections like "The Rickies"
and "The Flexies"). Bare `conctl rickies` (no `--json`) leads with a one-line
bill banner above the standings; `conctl rickies --json` still returns just the
standings (winners).
`rickies game <name>` matches the name fuzzily (unique substring is enough) and
returns the full game: `{ name, game-type, date-picked, date-graded, main-game, the-flexies }`.
Each round (`main-game`, `the-flexies`) has `picks[]` (each with `host`, `text`,
`score`, `pick-conditions` — a pick hit if `score > 0`), `scores[]`
(main: `{host, score}`; flexies: `{host, correct, total}`), and a `winner`.
`rickies --json` (keys are lowercase):
```json
{ "titles": { "Stephen": ["Annual Chairman", "Attorney General Flexie"], "Myke": [], "Federico": ["…"] },
  "annual":  { "winner": "Stephen", "game-name": "2025 Annual Rickies", "date": "2026-01-08", "title": "Annual Chairman", "social": "…" },
  "keynote": { "winner": "Federico", "game-name": "2026 March Rickies", "date": "2026-03-05", "title": "Keynote Chairman", "social": "…" },
  "flexies": { "…": "…" }, "annual-flexies": { "…": "…" } }
```

## Important behaviors

- **Transcripts lag the feed.** Transcript-backed output (search, and the
  default/`--say`/`--llm` modes of `--intro`/`--exit`) only covers episodes David
  Smith has transcribed — usually a few behind the newest. For a defaulted
  (not explicitly chosen) latest episode, `--intro`/`--exit` automatically fall
  back to the newest transcribed episode and print a note on **stderr**. If you
  explicitly pass `--episode N` for an untranscribed episode, the command errors
  (stderr) telling you the newest transcribed number — retry with that.
- **Chapters and `--play` never lag** — they come from the MP3, so they work for
  any episode including the newest.
- **`--play` requires `ffmpeg`**; everything else is dependency-free except
  **`--llm`, which requires `claude` or `codex` on PATH**. Both fail with a clear
  stderr message if the tool is missing.
- **Read-only.** conctl never modifies anything.

## Recipes

```bash
# When did they first talk about the Vision Pro, and link me to the moment?
conctl search "vision pro" --limit 1 --json | jq '.[0] | {episodeNumber, time, url}'

# Summarize how the latest *transcribed* episode ended.
conctl exit --llm --prompt conclusion

# Who is the current Annual Chairman?
conctl rickies --json | jq -r '.annual.winner'

# All links from a given episode's show notes.
conctl notes 600 --json | jq -r '.links[] | "\(.text)\t\(.url)"'
```
