package lang

import (
	"bufio"
	"os"
	"strings"
)

// parseIniFile is a minimal stand-in for PHP's parse_ini_file(..., false, INI_SCANNER_RAW).
//
// PocketMine's language files are flat `key = value` pairs with no sections, so this only
// handles that shape — not general INI (sections, arrays, typed scanning) — which is all real
// translation files ever use.
func parseIniFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		result[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// stripCSlashes approximates PHP's stripcslashes(): unescapes the handful of C-style sequences
// translation files actually use (\n, \t, \r, \\, etc). PHP's version also handles octal/hex
// byte escapes (\0-\7, \x..), which real translation files don't use, so those are left as-is.
func stripCSlashes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case 'v':
				b.WriteByte('\v')
			case 'f':
				b.WriteByte('\f')
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
