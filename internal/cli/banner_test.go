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
	if !strings.Contains(got, "the Connected CLI") {
		t.Errorf("banner tagline missing:\n%s", got)
	}
	if !strings.Contains(got, "██████") {
		t.Errorf("banner wordmark missing:\n%s", got)
	}
	// No-arg must NOT dump the full help rundown.
	if strings.Contains(got, "Tonight's rundown") {
		t.Errorf("no-arg should not print full help content:\n%s", got)
	}
}
