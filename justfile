# Default: list recipes
default:
    @just --list

# Build the conctl binary into ./bin
build:
    go build -o bin/conctl ./cmd/conctl

# Run conctl (e.g. just run search "vision pro")
run *args:
    go run ./cmd/conctl {{args}}

# Run all tests
test:
    go test ./...

# Remove build artifacts
kill:
    -rm -rf bin dist

# Cut a release (builds, signs, notarizes, publishes, updates the cask)
release:
    ./scripts/release.sh
