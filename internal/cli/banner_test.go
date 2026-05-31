package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestNoArgShowsBannerNotHelp(t *testing.T) {
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "The Connected CLI") {
		t.Errorf("banner tagline missing:\n%s", got)
	}
	if !strings.Contains(got, "██████") {
		t.Errorf("banner wordmark missing:\n%s", got)
	}
	// No-arg includes the rundown...
	if !strings.Contains(got, "Chapters:") {
		t.Errorf("no-arg should include the Chapters rundown:\n%s", got)
	}
	// ...but not cobra's usage/flags help.
	if strings.Contains(got, "Usage:") || strings.Contains(got, "Flags:") {
		t.Errorf("no-arg should not print cobra usage/flags:\n%s", got)
	}
	// The --json line was removed from both modes.
	if strings.Contains(got, "Add --json") {
		t.Errorf("the 'Add --json' line should be gone:\n%s", got)
	}
}
