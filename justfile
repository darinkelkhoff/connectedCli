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

# edit the files to do a release
pre-release:
   vi internal/cli/versions.txt
   vi internal/cli/root.go
   git commit internal/cli/versions.txt internal/cli/root.go

# Cut a release (builds, signs, notarizes, publishes, updates the cask)
release:
    ./scripts/release.sh
