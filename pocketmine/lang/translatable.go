package lang

import (
	"fmt"
	"strconv"
)

// Translatable is a port of pocketmine\lang\Translatable.
//
// Params are positional (index 0, 1, 2, ... matching the {%0}, {%1}, ... placeholder syntax
// almost every real translation string uses). PHP technically allows string-keyed params too;
// that's simplified away here since it's essentially unused in practice.
type Translatable struct {
	text   string
	params []any // each element is a string or *Translatable
	// keys are the parameter keys (PHP's array keys): nil for positional parameters 0..n-1, or one
	// key per parameter for translations with named placeholders like {%error}.
	keys []string
}

// NewTranslatable mirrors the Translatable constructor: non-Translatable params are stringified
// immediately (matching PHP's `(string) $param` cast at construction time, not at translate time).
func NewTranslatable(text string, params []any) *Translatable {
	return &Translatable{text: text, params: stringifyParams(params)}
}

// NewNamedTranslatable is the Translatable constructor with string-keyed parameters
// (`new Translatable($text, ["error" => $error, ...])`): keys[i] names params[i].
func NewNamedTranslatable(text string, keys []string, params []any) *Translatable {
	if len(keys) != len(params) {
		panic("lang: every named translation parameter needs a key")
	}
	return &Translatable{text: text, params: stringifyParams(params), keys: append([]string(nil), keys...)}
}

func stringifyParams(params []any) []any {
	p := make([]any, len(params))
	for i, param := range params {
		if t, ok := param.(*Translatable); ok {
			p[i] = t
		} else {
			p[i] = fmt.Sprintf("%v", param)
		}
	}
	return p
}

func (t *Translatable) Text() string      { return t.text }
func (t *Translatable) Parameters() []any { return t.params }

// ParameterKey returns the key of the i-th parameter: its name for named parameters, otherwise
// its index.
func (t *Translatable) ParameterKey(i int) string {
	if t.keys != nil {
		return t.keys[i]
	}
	return strconv.Itoa(i)
}

// Parameter is Translatable::getParameter for an integer key.
func (t *Translatable) Parameter(i int) any {
	if t.keys != nil {
		return t.NamedParameter(strconv.Itoa(i))
	}
	if i < 0 || i >= len(t.params) {
		return nil
	}
	return t.params[i]
}

// NamedParameter is Translatable::getParameter for a string key.
func (t *Translatable) NamedParameter(key string) any {
	for i := range t.params {
		if t.ParameterKey(i) == key {
			return t.params[i]
		}
	}
	return nil
}

func (t *Translatable) withText(text string) *Translatable {
	return &Translatable{text: text, params: t.params, keys: t.keys}
}

func (t *Translatable) Format(before, after string) *Translatable {
	return t.withText(before + "%" + t.text + after)
}

func (t *Translatable) Prefix(prefix string) *Translatable {
	return t.withText(prefix + "%" + t.text)
}

// Postfix mirrors Translatable::postfix() — which, in the PHP original, constructs its result
// without carrying the original params over (format()/prefix() do; postfix() doesn't). That
// looks like an oversight in the original, but this preserves it exactly rather than guessing.
func (t *Translatable) Postfix(postfix string) *Translatable {
	return NewTranslatable("%"+t.text+postfix, nil)
}
