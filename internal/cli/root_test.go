package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpMentionsRundown(t *testing.T) {
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "The Rickies") {
		t.Fatalf("help should read like a show rundown, got:\n%s", out.String())
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
