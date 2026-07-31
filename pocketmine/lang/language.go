package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const FallbackLanguage = "eng"

// LanguageNameKey is the translation key every language file must define, giving its own
// human-readable name (e.g. "English"). Pulled out of the generated KnownTranslationKeys table
// (deferred — see the package doc) since Language itself depends on this one specific key.
const LanguageNameKey = "language.name"

// Language is a port of pocketmine\lang\Language.
type Language struct {
	langName     string
	lang         map[string]string
	fallbackLang map[string]string
}

// NewLanguage loads langCode.ini (and fallback.ini) from path.
func NewLanguage(langCode string, path string, fallback string) (*Language, error) {
	if fallback == "" {
		fallback = FallbackLanguage
	}
	l := &Language{langName: strings.ToLower(langCode)}

	loaded, err := loadLang(path, l.langName)
	if err != nil {
		return nil, err
	}
	l.lang = loaded

	loadedFallback, err := loadLang(path, fallback)
	if err != nil {
		return nil, err
	}
	l.fallbackLang = loadedFallback

	return l, nil
}

// GetLanguageList scans path for `<code>.ini` files and returns code -> display name, using each
// file's own LanguageNameKey entry.
func GetLanguageList(path string) (map[string]string, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, &LanguageNotFoundException{Message: fmt.Sprintf("Language directory %s does not exist or is not a directory", path)}
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, &LanguageNotFoundException{Message: fmt.Sprintf("Language directory %s does not exist or is not a directory", path)}
	}

	result := map[string]string{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".ini") {
			continue
		}
		code := strings.SplitN(name, ".", 2)[0]
		strs, err := loadLang(path, code)
		if err != nil {
			continue
		}
		if v, ok := strs[LanguageNameKey]; ok {
			result[code] = v
		}
	}
	return result, nil
}

func loadLang(path string, languageCode string) (map[string]string, error) {
	file := filepath.Join(path, languageCode+".ini")
	if _, err := os.Stat(file); err == nil {
		raw, err := parseIniFile(file)
		if err == nil && len(raw) > 0 {
			result := make(map[string]string, len(raw))
			for k, v := range raw {
				result[k] = stripCSlashes(v)
			}
			return result, nil
		}
	}
	return nil, &LanguageNotFoundException{Message: fmt.Sprintf("Language %q not found", languageCode)}
}

func (l *Language) Name() string { return l.Get(LanguageNameKey) }
func (l *Language) Lang() string { return l.langName }

func (l *Language) internalGet(id string) (string, bool) {
	if v, ok := l.lang[id]; ok {
		return v, true
	}
	if v, ok := l.fallbackLang[id]; ok {
		return v, true
	}
	return "", false
}

// Get returns the translated string for id, or id itself if no translation exists.
func (l *Language) Get(id string) string {
	if v, ok := l.internalGet(id); ok {
		return v
	}
	return id
}

func (l *Language) GetAll() map[string]string {
	result := make(map[string]string, len(l.lang))
	for k, v := range l.lang {
		result[k] = v
	}
	return result
}

func (l *Language) getUsedParameterCount(rawString string, given int) int {
	highestIndex := -1
	for i := 0; i < given; i++ {
		if strings.Contains(rawString, fmt.Sprintf("{%%%d}", i)) {
			highestIndex = i
		}
	}
	return highestIndex + 1
}

// TranslateString is a port of Language::translateString(). onlyPrefix, if non-nil, limits
// substitution to translation keys with that prefix (used to let a Bedrock client do its own
// translating of vanilla strings); untranslatedParameterCount is returned instead of taken by
// reference, since Go doesn't have PHP's inout parameters.
func (l *Language) TranslateString(str string, params []any, onlyPrefix *string) (result string, untranslatedParameterCount int) {
	baseText, found := l.internalGet(str)
	parameterCount := len(params)

	if found {
		if onlyPrefix != nil && !strings.HasPrefix(str, *onlyPrefix) {
			return str, l.getUsedParameterCount(baseText, parameterCount)
		}
	} else {
		baseText, parameterCount = l.parseTranslation(str, onlyPrefix, parameterCount)
		untranslatedParameterCount = parameterCount
	}

	for i, p := range params {
		baseText = strings.ReplaceAll(baseText, fmt.Sprintf("{%%%d}", i), stringifyParam(l, p))
	}

	return baseText, untranslatedParameterCount
}

// Translate is a port of Language::translate(): fully resolves a Translatable (and its nested
// Translatable parameters) into a plain string.
func (l *Language) Translate(t *Translatable) string {
	baseText, found := l.internalGet(t.Text())
	if !found {
		baseText, _ = l.parseTranslation(t.Text(), nil, 0)
	}

	for i, p := range t.Parameters() {
		baseText = strings.ReplaceAll(baseText, fmt.Sprintf("{%%%d}", i), stringifyParam(l, p))
	}

	return baseText
}

func stringifyParam(l *Language, p any) string {
	if t, ok := p.(*Translatable); ok {
		return l.Translate(t)
	}
	if s, ok := p.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", p)
}

func (l *Language) replaceTranslationKey(replaceString string, onlyPrefix *string, untranslatedParameterCount int, givenParameterCount int) (string, int) {
	key := replaceString[1:] // strip the leading "%"
	if t, ok := l.internalGet(key); ok {
		if onlyPrefix != nil && !strings.HasPrefix(key, *onlyPrefix) {
			used := l.getUsedParameterCount(t, givenParameterCount)
			if used < untranslatedParameterCount {
				used = untranslatedParameterCount
			}
			return replaceString, used
		}
		return t, untranslatedParameterCount
	}
	return replaceString, givenParameterCount
}

// parseTranslation replaces "%translation.key" tokens embedded inside text with their raw
// values, character by character (matching the PHP original's hand-rolled scanner: a translation
// key is a run of [0-9A-Za-z.-] immediately following a "%").
func (l *Language) parseTranslation(text string, onlyPrefix *string, givenParameterCount int) (string, int) {
	untranslatedParameterCount := 0
	var newString strings.Builder
	var replaceString *string

	flush := func(c byte) {
		replaced, count := l.replaceTranslationKey(*replaceString, onlyPrefix, untranslatedParameterCount, givenParameterCount)
		newString.WriteString(replaced)
		untranslatedParameterCount = count
		replaceString = nil
		if c == '%' {
			s := "%"
			replaceString = &s
		} else {
			newString.WriteByte(c)
		}
	}

	for i := 0; i < len(text); i++ {
		c := text[i]
		if replaceString != nil {
			if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '.' || c == '-' {
				*replaceString += string(c)
			} else {
				flush(c)
			}
		} else if c == '%' {
			s := "%"
			replaceString = &s
		} else {
			newString.WriteByte(c)
		}
	}
	if replaceString != nil {
		replaced, count := l.replaceTranslationKey(*replaceString, onlyPrefix, untranslatedParameterCount, givenParameterCount)
		newString.WriteString(replaced)
		untranslatedParameterCount = count
	}

	return newString.String(), untranslatedParameterCount
}
