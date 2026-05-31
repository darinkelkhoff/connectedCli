package cli

import "testing"

func TestFormatTitles(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, "Man of the People"},
		{[]string{}, "Man of the People"},
		{[]string{"Annual Chairman"}, "Annual Chairman"},
		{[]string{"Annual Chairman", "Attorney General Flexie"}, "Annual Chairman; Attorney General Flexie"},
	}
	for _, c := range cases {
		if got := formatTitles(c.in); got != c.want {
			t.Errorf("formatTitles(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
