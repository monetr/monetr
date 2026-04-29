package validators

import (
	"unicode"

	"github.com/monetr/validation"
)

// ErrPrintableUnicode is what [PrintableUnicode] returns when the value has any
// rune that isn't printable.
var ErrPrintableUnicode = validation.NewError(
	"validation_is_printable_unicode",
	"must contain printable characters only",
)

// PrintableUnicode is the same idea as [is.PrintableASCII] but doesn't reject
// every non-ASCII rune. It accepts whatever Go's [unicode.IsPrint] thinks is
// printable, which means letters, marks, numbers, punctuation, symbols, and the
// regular ASCII space. monetr uses this for free text fields where users might
// reasonably type things like "Pösted" or "Café" or even an emoji, but where we
// still don't want the value to contain a tab or a newline or one of those
// invisible whitespace characters that sneak in from copy/paste (no break
// space, zero width joiner, etc).
//
// Empty values pass, the same way the rest of the [validation.StringRule] rules
// do. Pair it with [validation.Required] when the field is required.
//
// One thing to know about: invalid UTF-8 byte sequences get replaced with
// U+FFFD when Go ranges over the string, and U+FFFD itself is printable. So
// this rule won't catch malformed input on its own. If we ever care about that
// we'll have to add a utf8.ValidString check at the boundary; so far it hasn't
// come up.
var PrintableUnicode = validation.NewStringRuleWithError(
	func(s string) bool {
		for _, r := range s {
			if !unicode.IsPrint(r) {
				return false
			}
		}
		return true
	},
	ErrPrintableUnicode,
)
