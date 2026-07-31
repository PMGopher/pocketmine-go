package lang

import "fmt"

// Translatable is a port of pocketmine\lang\Translatable.
//
// Params are positional (index 0, 1, 2, ... matching the {%0}, {%1}, ... placeholder syntax
// almost every real translation string uses). PHP technically allows string-keyed params too;
// that's simplified away here since it's essentially unused in practice.
type Translatable struct {
	text   string
	params []any // each element is a string or *Translatable
}

// NewTranslatable mirrors the Translatable constructor: non-Translatable params are stringified
// immediately (matching PHP's `(string) $param` cast at construction time, not at translate time).
func NewTranslatable(text string, params []any) *Translatable {
	p := make([]any, len(params))
	for i, param := range params {
		if t, ok := param.(*Translatable); ok {
			p[i] = t
		} else {
			p[i] = fmt.Sprintf("%v", param)
		}
	}
	return &Translatable{text: text, params: p}
}

func (t *Translatable) Text() string      { return t.text }
func (t *Translatable) Parameters() []any { return t.params }

func (t *Translatable) Parameter(i int) any {
	if i < 0 || i >= len(t.params) {
		return nil
	}
	return t.params[i]
}

func (t *Translatable) Format(before, after string) *Translatable {
	return NewTranslatable(before+"%"+t.text+after, t.params)
}

func (t *Translatable) Prefix(prefix string) *Translatable {
	return NewTranslatable(prefix+"%"+t.text, t.params)
}

// Postfix mirrors Translatable::postfix() — which, in the PHP original, constructs its result
// without carrying the original params over (format()/prefix() do; postfix() doesn't). That
// looks like an oversight in the original, but this preserves it exactly rather than guessing.
func (t *Translatable) Postfix(postfix string) *Translatable {
	return NewTranslatable("%"+t.text+postfix, nil)
}
