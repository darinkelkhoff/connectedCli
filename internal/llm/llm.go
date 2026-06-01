package llm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Provider is one local AI backend.
type Provider interface {
	Name() string
	Available() bool
	Generate(ctx context.Context, system, user string) (string, error)
}

// cliProvider shells out to an installed agent CLI in print/exec mode.
type cliProvider struct {
	name string
	bin  string
	args func(prompt string) []string
}

func (p cliProvider) Name() string   { return p.name }
func (p cliProvider) Available() bool { _, err := exec.LookPath(p.bin); return err == nil }

func (p cliProvider) Generate(ctx context.Context, system, user string) (string, error) {
	prompt := user
	if system != "" {
		prompt = system + "\n\n" + user
	}
	cmd := exec.CommandContext(ctx, p.bin, p.args(prompt)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w: %s", p.name, err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// DefaultProviders is the detection priority order.
func DefaultProviders() []Provider {
	return []Provider{
		cliProvider{name: "claude", bin: "claude", args: func(p string) []string {
			return []string{"-p", p}
		}},
		cliProvider{name: "codex", bin: "codex", args: func(p string) []string {
			return []string{"exec", p}
		}},
		cliProvider{name: "opencode", bin: "opencode", args: func(p string) []string {
			return []string{"run", p}
		}},
	}
}

func selectProvider(providers []Provider, force string) (Provider, error) {
	for _, p := range providers {
		if force != "" && p.Name() != force {
			continue
		}
		if p.Available() {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no local AI CLI found (looked for: claude, codex, opencode). Install one, or skip --llm")
}

// Generate runs a preset prompt against the first available provider.
// Returns the generated text and the provider name used.
func Generate(ctx context.Context, presetName, transcript, force string) (string, string, error) {
	preset, ok := Preset(presetName)
	if !ok {
		return "", "", fmt.Errorf("unknown prompt preset %q (try: %s)", presetName, strings.Join(PresetNames(), ", "))
	}
	p, err := selectProvider(DefaultProviders(), force)
	if err != nil {
		return "", "", err
	}
	user := preset.User + "\n\n---\n" + transcript
	text, err := p.Generate(ctx, preset.System, user)
	return text, p.Name(), err
}
