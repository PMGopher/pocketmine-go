package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/utils"
)

// FormattedCommandAlias is a port of pocketmine\command\FormattedCommandAlias: used to register
// commands defined in the `aliases` section of pocketmine.yml.
type FormattedCommandAlias struct {
	Command
	formatStrings []string
}

func NewFormattedCommandAlias(alias string, formatStrings []string) *FormattedCommandAlias {
	return &FormattedCommandAlias{
		Command:       InitCommand(alias, "User defined command", nil, nil),
		formatStrings: formatStrings,
	}
}

// formatStringRegex matches a placeholder like $1, $$1, $1-, $$1- at the start of the remaining
// string. PHP's equivalent uses `\G` (anchor at a given offset) plus a negative lookahead to
// forbid a leading zero; Go's RE2 engine supports neither \G nor lookaround, but both are
// expressible without them: matching against formatString[offset:] with a `^`-anchored pattern
// is exactly what \G at that offset means, and `[1-9][0-9]*` (first digit non-zero) says exactly
// what `(?!0)+\d+` was trying to.
var formatStringRegex = regexp.MustCompile(`^\$(\$)?([1-9][0-9]*)(-)?`)

func extractPlaceholderInfo(commandString string, offset int) (fullPlaceholder string, required bool, position int, variadic bool, ok bool) {
	m := formatStringRegex.FindStringSubmatch(commandString[offset:])
	if m == nil {
		return "", false, 0, false, false
	}
	position, _ = strconv.Atoi(m[2])
	return m[0], m[1] != "", position, m[3] != "", true
}

func buildReplacement(args []string, position int, rest bool) (string, bool) {
	if rest && position < len(args) {
		return strings.Join(args[position:], " "), true
	}
	if position < len(args) {
		return args[position], true
	}
	return "", false
}

// buildCommand is a port of FormattedCommandAlias::buildCommand(). unresolved reports a missing
// *optional* placeholder (PHP's null return); err reports an actual problem (invalid token or a
// missing *required* placeholder).
func buildCommand(formatString string, args []string) (result string, unresolved bool, err error) {
	index := 0
	for {
		pos := strings.IndexByte(formatString[index:], '$')
		if pos == -1 {
			break
		}
		start := index + pos

		if start > 0 && formatString[start-1] == '\\' {
			formatString = formatString[:start-1] + formatString[start:]
			// index is deliberately left unchanged: the string just got 1 byte shorter, so
			// searching for '$' again from the same numeric offset in the shrunk string correctly
			// resumes just past the '$' we unescaped, rather than re-matching it as a placeholder.
			continue
		}

		fullPlaceholder, required, position, variadic, ok := extractPlaceholderInfo(formatString, start)
		if !ok {
			return "", false, fmt.Errorf("invalid replacement token")
		}
		position-- // placeholders are 1-based; args is 0-based

		if required && position >= len(args) {
			return "", false, fmt.Errorf("missing required argument %d", position+1)
		}

		replacement, has := buildReplacement(args, position, variadic)
		if !has {
			return "", true, nil
		}

		end := start + len(fullPlaceholder)
		formatString = formatString[:start] + replacement + formatString[end:]
		index = start + len(replacement)
	}
	return formatString, false, nil
}

func (c *FormattedCommandAlias) Execute(sender Sender, commandLabel string, args []string) (any, error) {
	var commandsToRun [][]string

	for _, formatString := range c.formatStrings {
		formatArgs := ParseQuoteAware(formatString)
		var unresolvedList []string
		var processedArgs []string
		for _, formatArg := range formatArgs {
			processed, unresolved, err := buildCommand(formatArg, args)
			if err != nil {
				sender.SendMessage(utils.Red + err.Error())
				return false, nil
			}
			if unresolved {
				unresolvedList = append(unresolvedList, formatArg)
			} else if len(unresolvedList) != 0 {
				// unresolved args are OK only if they're at the end — we can't have holes in the args list
				sender.SendMessage(fmt.Sprintf("%sUnable to resolve format arguments (%s) in command string %q due to missing arguments",
					utils.Red, strings.Join(unresolvedList, ", "), formatString))
				return false, nil
			} else {
				processedArgs = append(processedArgs, processed)
			}
		}
		commandsToRun = append(commandsToRun, processedArgs)
	}

	result := true
	commandMap := sender.GetServer().GetCommandMap()
	for _, cmdArgs := range commandsToRun {
		if len(cmdArgs) == 0 {
			panic("This should have been checked before construction")
		}
		label, rest := cmdArgs[0], cmdArgs[1:]

		// This approximately duplicates SimpleCommandMap.Dispatch's logic, to invoke the target
		// directly with pre-parsed arguments rather than rebuilding and re-parsing a command
		// string for no reason. Note this does NOT call TestPermission, matching the PHP original.
		if target := commandMap.GetCommand(label); target != nil {
			_, err := target.Execute(sender, label, rest)
			if _, ok := err.(*InvalidCommandSyntaxException); ok {
				sender.SendMessage(fmt.Sprintf("Usage: %s", stringifyMessage(target.Usage())))
			}
		} else {
			sender.SendMessage(fmt.Sprintf("Unknown command: %q. Type \"/help\" for help.", label))
			// To match the behaviour of SimpleCommandMap.Dispatch. This shouldn't normally happen,
			// but might if the command was unregistered or modified after the alias was installed.
			result = false
		}
	}

	return result, nil
}
