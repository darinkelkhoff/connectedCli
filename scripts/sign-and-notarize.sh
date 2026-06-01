#!/usr/bin/env bash
#
# Sign and notarize a single conctl binary for distribution outside the
# Mac App Store. Invoked by GoReleaser's per-build post hook, once per
# target binary (darwin/amd64 and darwin/arm64).
#
# Prerequisites (one-time, on the release machine):
#   1. A "Developer ID Application" certificate in the login keychain.
#   2. A stored notarytool credential profile. Create it with:
#        xcrun notarytool store-credentials conctl-notary \
#          --apple-id "you@example.com" \
#          --team-id 9MG4YT2G93 \
#          --password "<app-specific-password>"
#      (App-specific passwords are created at appleid.apple.com.)
#
# Overridable via env:
#   CONCTL_SIGN_IDENTITY  - codesign identity (default: the Developer ID below)
#   CONCTL_NOTARY_PROFILE - notarytool keychain profile name (default: conctl-notary)

set -euo pipefail

BINARY="${1:?usage: sign-and-notarize.sh <binary-path>}"

# Snapshot rehearsals (`goreleaser release --snapshot`) still run build hooks,
# but there's no point signing/notarizing a throwaway build. Set
# CONCTL_SKIP_SIGN=1 to no-op.
if [[ -n "${CONCTL_SKIP_SIGN:-}" ]]; then
  echo "==> CONCTL_SKIP_SIGN set; skipping sign+notarize for $BINARY"
  exit 0
fi

IDENTITY="${CONCTL_SIGN_IDENTITY:-Developer ID Application: Darin Kelkhoff (9MG4YT2G93)}"
NOTARY_PROFILE="${CONCTL_NOTARY_PROFILE:-conctl-notary}"

echo "==> Signing $BINARY"
# --options runtime  : enable the hardened runtime (required for notarization)
# --timestamp        : embed a secure timestamp (required for notarization)
# --force            : re-sign if already signed
codesign --force --options runtime --timestamp \
  --sign "$IDENTITY" "$BINARY"

echo "==> Verifying signature"
codesign --verify --strict --verbose=2 "$BINARY"

# notarytool cannot accept a bare executable; it must be wrapped in a zip
# (or pkg/dmg). A bare binary also can't be "stapled", so we don't staple —
# once Apple accepts it, Gatekeeper validates online via the signature hash.
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT
ZIP="$WORKDIR/$(basename "$BINARY").zip"

echo "==> Zipping for notarization: $ZIP"
ditto -c -k "$BINARY" "$ZIP"

echo "==> Submitting to Apple notary service (this can take a minute)"
xcrun notarytool submit "$ZIP" \
  --keychain-profile "$NOTARY_PROFILE" \
  --wait

echo "==> Done: $BINARY is signed and notarized"
