package render

import (
	"encoding/json"
	"io"
)

// JSON writes v as indented JSON followed by a newline.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ErrorJSON writes a structured error object.
func ErrorJSON(w io.Writer, msg string) {
	_ = JSON(w, map[string]string{"error": msg})
}
