# conctl — Development

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

Releases are cut locally with `scripts/release.sh` (wrapped by `just release`). Versions are **episode-numbered** — `001`, `002`, … — not semver. Each macOS binary is signed with a Developer ID certificate and notarized by Apple, so it passes Gatekeeper without any quarantine workarounds.

The version is the latest entry in `internal/cli/versions.txt`; the release script reads it from there and refuses to publish if the built binary's self-reported version doesn't match.

### One-time setup

1. **Developer ID Application certificate** in the login keychain (verify with `security find-identity -v -p codesigning`).
2. **Notary credentials** stored under the profile name the signing script expects (`conctl-notary`):
   ```bash
   xcrun notarytool store-credentials conctl-notary \
     --apple-id "you@example.com" --team-id 9MG4YT2G93
   ```
   (Uses an app-specific password from appleid.apple.com.)
3. **`gh` authenticated** with push access to both `connectedCli` and
   `homebrew-tap`. The release uses your existing `gh`/git credentials — no
   separate token to export.

### Creating a release

1. Add the new episode line to `internal/cli/versions.txt` (e.g. `002 <codename>`) and bump `Version` in `internal/cli/root.go` to match.
2. Commit, then from a clean tree:
   ```bash
   just release
   ```

The script builds the darwin amd64/arm64 binaries, runs `scripts/sign-and-notarize.sh` on each (sign → notarize → wait; a few minutes total), tags the commit (`002`), publishes the archives + checksums to a GitHub Release, and writes the cask to `darinkelkhoff/homebrew-tap`. Users then `brew install darinkelkhoff/tap/conctl`.

Set `CONCTL_SKIP_SIGN=1` to build and publish without signing/notarizing (testing only — the result won't pass Gatekeeper).

If notarization fails, inspect the reason with:
```bash
xcrun notarytool log <submission-id> --keychain-profile conctl-notary
```
