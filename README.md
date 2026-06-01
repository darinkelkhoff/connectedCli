# conctl: The Connected CLI

A command-line companion app for the [Connected](https://relay.fm/connected) podcast.  Built for fans, hosts, AI agents, and mostly just because Stephen mentioned it as a joke once.

```text
 ██████╗   ██████╗   ███╗   ██╗   ██████╗  ████████╗  ██╗
██╔════╝  ██╔═══██╗  ████╗  ██║  ██╔════╝  ╚══██╔══╝  ██║
██║       ██║   ██║  ██╔██╗ ██║  ██║          ██║     ██║
██║       ██║   ██║  ██║╚██╗██║  ██║          ██║     ██║
██║       ██║   ██║  ██║╚██╗██║  ██║          ██║     ██║
╚██████╗  ╚██████╔╝  ██║ ╚████║  ╚██████╗     ██║     ███████╗
 ╚═════╝   ╚═════╝   ╚═╝  ╚═══╝   ╚═════╝     ╚═╝     ╚══════╝
                     The Connected CLI
```

`conctl` can search the show's unofficial transcripts, print lists of chapters, dump show notes, track the Rickies, and more.  

It was inspired by Stephen saying "`connected --exit`" on [episode 605](https://relay.fm/connected/605), after hearing about [remctl](https://www.macstories.net/stories/introducing-remctl-the-power-user-reminders-cli-for-macos-and-ai-agents/). Like `remctl`, `conctl` produces human-readable terminal output by default, but can also produce JSON for agents to consume.  

## How It Works
`conctl` is a single, self-contained Go binary that stitches together the public data the show already publishes:

- episodes, chapters, and show notes come from the Connected RSS feed.
- transcript are from podcastsearch.david-smith.org (thank you David Smith)
- the Rickies data comes from rickies.net/api/v1 (thanks to jbiatek)
- local AI (via the `--llm` flag calls your your already-installed claude / codex CLI
  
Nothing is hosted by conctl, and there are no API keys. Chapters are read by fetching only the first ~512 KB of an episode's MP3 (the ID3 tag), so listing chapters is fast. Audio playback streams just the chosen chapter via `ffmpeg`.

## Quick Start

### Via Homebrew
```bash
brew install darin.kelkhoff/tap/conctl

conctl                       # print the show banner with a rundown of commands
conctl search "vision pro"   # search transcripts, deep-linked to the second
conctl rickies               # print current title holders
conctl --exit                # the sign-off (the final chapter of the latest episode)
```

The homebrew formula declares an optional `ffmpeg` dependency (only needed for `--play`).
### From source

```bash
git clone https://github.com/darin.kelkhoff/connectedCli.git
cd connectedCli
just build            # -> ./bin/conctl   (or: go build -o bin/conctl ./cmd/conctl)
./bin/conctl --help
```

## Command Map

| Task                    | Commands                                            |
| ----------------------- | --------------------------------------------------- |
| See the show            | `latest`, `chapters [ep]`, `notes [ep]`             |
| Search transcripts      | `search <query>`                                    |
| The Rickies             | `rickies`, `rickies titles [host]`, `rickies games` |
| Listen / read a chapter | `play [ep]`, `--intro`, `--exit`                    |
| Agent mode              | `--json` on any command                             |

Common examples:

```bash
conctl latest                          # most recent episode
conctl latest --full                   # episode + chapters + show notes in one view
conctl chapters                        # chapters of the latest episode
conctl chapters 605                    # chapters of a specific episode
conctl notes                           # show notes & links for the latest episode
conctl search "macintosh" --limit 5    # transcript hits, grouped by episode
conctl rickies                         # current Annual & Keynote Chairmen + titles
conctl rickies titles Federico         # one host's titles
conctl rickies games                   # every Rickies game ever
conctl play 605 --chapter 7            # play a chapter's audio (needs ffmpeg)
conctl intro                           # the cold open of the latest episode
conctl exit 601                        # the closing chapter of a specific episode
conctl exit --short                    # Arrivederci. Cheerio. Bye, y'all.
conctl exit --say                      # ...read aloud by macOS
conctl exit --llm --prompt haiku       # ...as a haiku, from your local AI
conctl --exit                          # the joke: works as a flag too
conctl rickies --json                  # structured output for agents
```

## Chapters, Intros & Exits

Every episode's chapters come from the MP3's embedded ID3 chapter markers.
conctl treats the **first** chapter as the intro and the **last** as the exit:

```bash
conctl chapters 605      # list them
conctl intro             # first chapter of the latest episode
conctl exit              # last chapter of the latest episode
conctl play --last       # play the last chapter (ffmpeg)
```

The `intro` and `exit` subcommands share a set of render modes (choose at most one):

| flag        | what you get                                                        |
| ----------- | ------------------------------------------------------------------- |
| *(default)* | the chapter's transcript text — no dependencies                     |
| `--play`    | streams just that chapter's audio (requires `ffmpeg`)               |
| `--say`     | macOS `say` reads the chapter aloud                                 |
| `--llm`     | a local AI CLI (`claude` / `codex`) generates text from the chapter |
| `--short`   | just the quick sign-offs                                            |
| `--json`    | structured output - ideal for consumption by an LLM                 |

Target any specific episode with a positional number (`conctl intro 601`); the
default is the latest episode (*latest transcribed episode, when requesting text
instead of audio*). For the joke, `conctl --intro` and `conctl --exit` also work
as flags on the bare command.

## Search

`conctl search <query>` runs against David Smith's Whisper transcripts and groups results by episode, with the matched term highlighted. In a terminal, each timestamp is a clickable link to that exact second on the transcript site:

```text
#601 I Love Wrists:
  01:40:25  … the best thing you can say about the Vision Pro, I think. …
  01:40:19  … Tim Cook says about the Vision Pro, it's tomorrow's engineering …
```

Add `--json` for `{episodeNumber, episodeTitle, time, seconds, snippet, url}` objects.

## The Rickies

conctl reads the [rickies.net](https://rickies.net/) API:

```bash
conctl rickies                 # a Bill-of-Rickies banner + current chairmen & titles
conctl rickies bill            # the full current Bill of Rickies (the rules)
conctl rickies bill --versions # list every past edition of the bill (since 2017)
conctl rickies bill --version annual-2017   # print a past edition
conctl rickies titles Stephen  # one host's titles
conctl rickies games           # every Rickies game, 2017 -> today
conctl rickies game "2026 March"   # one game in detail
```

`rickies bill` reconstructs the **current** rules from the date-versioned
[Bill of Rickies](https://rickies.co/billof) — keeping only the rules (and inline
edits) in force today. `--versions` lists all 35 past editions (the page's
history slider), and `--version <slug|index>` prints any of them as they stood
at that date.

`rickies game <name>` shows the main round and the flexies for a game — the
winner, each host's score, and every pick grouped by host with a ✓/✗ for whether
it hit, plus the flexies' charity. The name is matched fuzzily (a unique
substring is enough).

A host with no current titles is, of course, a **Man of the People**.

## Show Notes

Each episode's notes — summary, sponsors, and links — live in the RSS feed, so
this works for every episode (no transcript needed):

```bash
conctl notes        # latest episode
conctl notes 600    # a specific episode
conctl notes --json # { episode, title, summary, links: [{text, url}] }
```

## Local AI (`--llm`)

`--llm` rewrites a chapter using an AI CLI **you already have installed** — no API
keys, no new services. It probes `claude` (`claude -p`), then `codex`
(`codex exec`), and uses the first one it finds.

```bash
conctl exit 601 --llm                    # a closing summary in the show's voice
conctl exit 601 --llm --prompt haiku     # the episode as a haiku
conctl intro 601 --llm --say             # an AI cold-open, read aloud
```

Prompt presets (`--prompt`): `conclusion` (default for `--exit`), `cold-open`
(default for `--intro`), `recap`, `style-federico`, `style-myke`,
`style-stephen`, `haiku`. A spinner shows while your local model thinks.

## For Agents

Every command accepts `--json` and writes machine-readable output to stdout;
notes, spinners, and progress go to stderr, so piping stdout stays clean.

```bash
conctl search "siri" --json | jq '.[0]'
conctl chapters 605 --json
conctl rickies --json
conctl notes --json
conctl exit 601 --json                # { episode, chapter, text }
```

Because conctl reads only public data and never writes anything, it's safe to
hand to an agent.

A ready-to-use agent guide — with the full `--json` schemas, caveats, and `jq`
recipes — lives in [`SKILL.md`](SKILL.md).

## Notes & Caveats

- **Transcripts lag the feed.** David Smith re-transcribes episodes on a delay,
  so the newest episode or two may not be searchable yet. For `--intro`/`--exit`
  text/`--say`/`--llm`, conctl automatically falls back to the newest transcribed
  episode (with a note on stderr). `--play` always works — chapters and audio
  come straight from the MP3.
- **`--play` needs `ffmpeg`** (`brew install ffmpeg`); the Homebrew formula
  declares it as an optional dependency. All other modes are dependency-free
  (`say` is built into macOS).
- **`--llm` needs a local agent CLI** (`claude` or `codex`); otherwise it prints
  a friendly "no local AI found" message.

## Development

```bash
just            # list recipes
just build      # build ./bin/conctl
just run ...    # go run with args (e.g. just run search "macintosh")
just test       # go test ./...
```

## Project Layout

| Path | Purpose |
| --- | --- |
| `cmd/conctl` | Entry point |
| `internal/cli` | Cobra command tree, banner, render modes, spinner |
| `internal/feed` | Connected RSS feed parsing (episodes, show notes) |
| `internal/chapters` | ID3v2 chapter parsing + ranged MP3 fetch |
| `internal/podsearch` | Transcript search + episode transcript slicing |
| `internal/rickies` | rickies.net API client |
| `internal/audio` | `ffplay` segment playback + macOS `say` |
| `internal/llm` | Local AI provider detection + prompt presets |
| `internal/render` | Pretty vs. JSON output |

## Releasing

```bash
git tag v0.1.0 && git push --tags
goreleaser release --clean      # needs GITHUB_TOKEN with repo scope
```

This publishes macOS binaries to a GitHub Release and updates the Homebrew
formula in the `dkelkhoff/homebrew-tap` repo. Users then
`brew install dkelkhoff/tap/conctl`.

## Credits

conctl stands on the shoulders of fan-built and host-built work:

- [Connected](https://relay.fm/connected) by Myke Hurley, Federico Viticci, and Stephen Hackett (Relay FM)
- [Podcast Search](https://podcastsearch.david-smith.org/) by David Smith
- [rickies.net](https://rickies.net/) by jbiatek
- Inspired by [remctl](https://github.com/viticci/remctl) by Federico Viticci

## License

MIT. See [LICENSE](LICENSE).
