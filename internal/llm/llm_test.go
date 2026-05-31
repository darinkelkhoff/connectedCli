package llm

import (
	"context"
	"testing"
)

type fakeProvider struct{ avail bool }

func (f fakeProvider) Name() string    { return "fake" }
func (f fakeProvider) Available() bool  { return f.avail }
func (f fakeProvider) Generate(_ context.Context, _, user string) (string, error) {
	return "GENERATED:" + user, nil
}

func TestSelectFirstAvailable(t *testing.T) {
	p, err := selectProvider([]Provider{fakeProvider{false}, fakeProvider{true}}, "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "fake" {
		t.Fatalf("got %s", p.Name())
	}
}

func TestSelectNoneAvailable(t *testing.T) {
	if _, err := selectProvider([]Provider{fakeProvider{false}}, ""); err == nil {
		t.Fatal("expected error when no provider available")
	}
}

func TestPresetLookup(t *testing.T) {
	p, ok := Preset("conclusion")
	if !ok || p.System == "" {
		t.Fatalf("conclusion preset missing: %+v ok=%v", p, ok)
	}
	if _, ok := Preset("nope"); ok {
		t.Fatal("unknown preset should miss")
	}
}
