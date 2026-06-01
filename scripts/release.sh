#!/usr/bin/env bash
#
# Release conctl: build, sign, notarize, publish a GitHub release, and update
# the Homebrew cask. Versions are episode-numbered (001, 002, ...) — no semver.
#
# The version is the latest entry in internal/cli/versions.txt. To cut a new
# release, add the new episode line there (and bump root.go's Version to match),
# commit, then run this script from a clean tree.
#
# Prereqs: a clean working tree, `gh` authenticated, push access to both repos,
# and the signing prereqs documented in scripts/sign-and-notarize.sh.
# Set CONCTL_SKIP_SIGN=1 to build/publish without signing (testing only).

set -euo pipefail

REPO="darinkelkhoff/connectedCli"
TAP_REMOTE="git@github.com:darinkelkhoff/homebrew-tap.git"
BINARY="conctl"

cd "$(dirname "$0")/.."  # repo root

# Version = first field of the last non-empty line of versions.txt (e.g. "001").
VERSION="$(awk 'NF{v=$1} END{print v}' internal/cli/versions.txt)"
[[ -n "$VERSION" ]] || { echo "could not read a version from internal/cli/versions.txt"; exit 1; }
echo "==> Releasing $BINARY $VERSION"

# Guard rails: clean tree, version not already released.
[[ -z "$(git status --porcelain)" ]] || { echo "working tree is dirty — commit first"; exit 1; }
if git rev-parse "refs/tags/$VERSION" >/dev/null 2>&1; then
  echo "tag $VERSION already exists locally"; exit 1
fi
if gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1; then
  echo "release $VERSION already exists on $REPO"; exit 1
fi

DIST="dist"
rm -rf "$DIST"; mkdir -p "$DIST"

assets=()
for arch in amd64 arm64; do
  out="$DIST/${BINARY}_${VERSION}_darwin_${arch}/${BINARY}"
  mkdir -p "$(dirname "$out")"
  echo "==> Building darwin/$arch"
  GOOS=darwin GOARCH="$arch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o "$out" ./cmd/conctl
  ./scripts/sign-and-notarize.sh "$out"
  tarball="$DIST/${BINARY}_${VERSION}_darwin_${arch}.tar.gz"
  tar -czf "$tarball" -C "$(dirname "$out")" "$BINARY"
  assets+=("$tarball")
done

# Sanity check: the binary's self-reported version must match versions.txt, so
# a release can't ship with root.go and versions.txt out of sync.
case "$(uname -m)" in
  arm64)  host_arch=arm64 ;;
  x86_64) host_arch=amd64 ;;
  *)      host_arch="" ;;
esac
if [[ -n "$host_arch" ]]; then
  reported="$("$DIST/${BINARY}_${VERSION}_darwin_${host_arch}/${BINARY}" --version | grep -oE '#[0-9]+' | tr -d '#' || true)"
  if [[ "$reported" != "$VERSION" ]]; then
    echo "binary reports version '#${reported}' but versions.txt latest is '${VERSION}'"
    echo "→ bump Version in internal/cli/root.go to match, commit, and retry"
    exit 1
  fi
fi

# Checksums.
( cd "$DIST" && shasum -a 256 "${BINARY}"_*.tar.gz > checksums.txt )
amd_sha="$(shasum -a 256 "$DIST/${BINARY}_${VERSION}_darwin_amd64.tar.gz" | awk '{print $1}')"
arm_sha="$(shasum -a 256 "$DIST/${BINARY}_${VERSION}_darwin_arm64.tar.gz" | awk '{print $1}')"

# Tag and publish the GitHub release.
echo "==> Tagging and publishing release $VERSION"
git tag "$VERSION"
git push origin "$VERSION"
gh release create "$VERSION" --repo "$REPO" \
  --title "conctl $VERSION" \
  --notes "conctl $VERSION" \
  "${assets[@]}" "$DIST/checksums.txt"

# Render the Homebrew cask and push it to the tap.
echo "==> Updating Homebrew cask in the tap"
tapdir="$(mktemp -d)"
trap 'rm -rf "$tapdir"' EXIT
git clone --depth 1 "$TAP_REMOTE" "$tapdir"
mkdir -p "$tapdir/Casks"
cat > "$tapdir/Casks/conctl.rb" <<CASK
cask "conctl" do
  version "$VERSION"

  on_macos do
    on_intel do
      sha256 "$amd_sha"
      url "https://github.com/$REPO/releases/download/#{version}/conctl_#{version}_darwin_amd64.tar.gz"
    end
    on_arm do
      sha256 "$arm_sha"
      url "https://github.com/$REPO/releases/download/#{version}/conctl_#{version}_darwin_arm64.tar.gz"
    end
  end

  name "conctl"
  desc "A command-line companion for the Connected podcast"
  homepage "https://github.com/$REPO"

  depends_on formula: "ffmpeg"

  binary "conctl"
end
CASK

git -C "$tapdir" add Casks/conctl.rb
git -C "$tapdir" commit -m "conctl $VERSION"
git -C "$tapdir" push

echo "==> Released conctl $VERSION"
echo "    Install: brew install darinkelkhoff/tap/conctl"
