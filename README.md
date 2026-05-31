# conctl — the Connected CLI

A command-line companion for the [Connected](https://relay.fm/connected) podcast,
built for both human power users and AI agents. Inspired by Stephen Hackett's
"connected cli… connected --exit" bit on episode 605 — itself a riff on
[remctl](https://www.macstories.net/stories/introducing-remctl-the-power-user-reminders-cli-for-macos-and-ai-agents/).

## Install

    brew install dkelkhoff/tap/conctl

## Usage

    conctl search "vision pro"     # search the transcripts, deep-linked to the second
    conctl rickies                 # current Rickies chairmen & titles
    conctl rickies games           # every Rickies game ever
    conctl chapters                # list the latest episode's chapters
    conctl play --last             # play a chapter (needs ffmpeg)
    conctl --intro                 # the cold open (first chapter)
    conctl --exit                  # the sign-off (last chapter)
    conctl --exit --say            # ...read aloud by macOS
    conctl --exit --llm            # ...rewritten by a local AI CLI (claude/codex)
    conctl --exit --short          # Arrivederci. Cheerio. Bye, y'all.

Add `--json` to any command for agent-friendly structured output.

Target a specific episode with `--episode N` (or a positional number on
`chapters`/`play`). Note: transcript text lags the feed by a few episodes
(David Smith re-transcribes on a delay), so text/`--say`/`--llm` on the very
latest episode may not be available yet — `--play` always works, since chapters
come straight from the MP3.

## Render modes (for --intro / --exit / play)

| flag       | output                                                            |
|------------|-------------------------------------------------------------------|
| *(default)*| the chapter's transcript text (no dependencies)                   |
| `--play`   | streams just that chapter's audio (requires `ffmpeg`)             |
| `--say`    | macOS `say` reads the chapter aloud                               |
| `--llm`    | a local AI CLI (`claude`/`codex`) generates from the chapter      |
| `--json`   | structured output                                                 |
| `--short`  | (intro/exit) just the quick sign-offs                            |

LLM prompt presets (`--prompt`): `conclusion`, `cold-open`, `recap`,
`style-federico`, `style-myke`, `style-stephen`, `haiku`.

## Data sources

- **Transcripts** — David Smith's [Podcast Search](https://podcastsearch.david-smith.org/) (Whisper).
- **The Rickies** — the [rickies.net](https://rickies.net/) JSON API (by jbiatek).
- **Episodes & chapters** — the Connected RSS feed; chapters parsed from the MP3's
  ID3 chapter frames via a ranged GET (no full download).

## Develop

    just build      # build ./bin/conctl
    just run ...    # go run with args
    just test       # go test ./...

## Releasing

1. Create the `dkelkhoff/homebrew-tap` GitHub repo (once).
2. Tag: `git tag v0.1.0 && git push --tags`.
3. `goreleaser release --clean` (needs `GITHUB_TOKEN` with repo scope).

This publishes macOS binaries to a GitHub Release and updates the Homebrew
formula in the tap. End users then `brew install dkelkhoff/tap/conctl`.
