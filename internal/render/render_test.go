package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONEmitsObject(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, map[string]int{"number": 605}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"number": 605`) {
		t.Fatalf("got %s", buf.String())
	}
}

func TestErrorJSON(t *testing.T) {
	var buf bytes.Buffer
	ErrorJSON(&buf, "boom")
	if !strings.Contains(buf.String(), `"error": "boom"`) {
		t.Fatalf("got %s", buf.String())
	}
}
