package command

import "regexp"

var quoteAwarePattern = regexp.MustCompile(`"((?:\\.|[^\\"])*)"|(\S+)`)
var escapedQuotePattern = regexp.MustCompile(`\\([\\"])`)

// ParseQuoteAware is a port of CommandStringHelper::parseQuoteAware(): splits a command line into
// arguments, treating unescaped-quoted sections as a single argument.
//
// Examples:
//   - `give "steve jobs" apple` -> ["give", "steve jobs", "apple"]
//   - `say "This is a \"string containing quotes\""` -> ["say", `This is a "string containing quotes"`]
func ParseQuoteAware(commandLine string) []string {
	var args []string
	for _, m := range quoteAwarePattern.FindAllStringSubmatch(commandLine, -1) {
		match := m[1]
		if match == "" {
			match = m[2]
		}
		args = append(args, escapedQuotePattern.ReplaceAllString(match, "$1"))
	}
	return args
}
