package cli

import "strconv"

// parseOptionalEp parses an optional positional episode number.
func parseOptionalEp(args []string) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}
	return strconv.Atoi(args[0])
}
