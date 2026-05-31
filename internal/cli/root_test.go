package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpIsUsageOnly(t *testing.T) {
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Usage:") {
		t.Fatalf("help should show usage, got:\n%s", got)
	}
	// The Chapters rundown belongs to the no-arg banner, not --help.
	if strings.Contains(got, "Chapters:") {
		t.Fatalf("help should NOT include the Chapters rundown, got:\n%s", got)
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "conctl") {
		t.Fatalf("version should mention conctl, got: %s", out.String())
	}
}
