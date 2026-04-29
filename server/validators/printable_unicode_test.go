package validators_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/validators"
	"github.com/stretchr/testify/assert"
)

func TestPrintableUnicode(t *testing.T) {
	// The rule leans on Go's [unicode.IsPrint], which is generous with
	// letters/marks/numbers/punctuation/symbols across scripts but strict about
	// anything else. The cases below try to pin both halves: the stuff we
	// actually want users to be able to type (umlauts, accents, CJK, emoji), and
	// the stuff we want to keep out (control characters, non-ASCII whitespace,
	// formatting code points). Empty input passes because that matches what the
	// other [validation.StringRule] rules do; we lean on [validation.Required]
	// elsewhere when the field has to be set.
	const wantErr = "must contain printable characters only"
	cases := []struct {
		name    string
		input   string
		wantErr string
	}{
		// Things we want to accept.
		{
			name:    "empty",
			input:   "",
			wantErr: "",
		},
		{
			name:    "ascii letters",
			input:   "POSTED",
			wantErr: "",
		},
		{
			name:    "ascii digits",
			input:   "1234567890",
			wantErr: "",
		},
		{
			name:    "ascii punctuation",
			input:   "!@#$%^&*()_+-=,.<>/?;:'\"[]{}|\\`~",
			wantErr: "",
		},
		{
			name:    "ascii space",
			input:   "hello world",
			wantErr: "",
		},
		{
			// The headline reason this rule exists. "Pösted" is the kind of value a
			// real bank export might use that the old ASCII rule would have rejected.
			name:    "umlauts",
			input:   "Pösted",
			wantErr: "",
		},
		{
			name:    "accented vowels",
			input:   "Café résumé naïve",
			wantErr: "",
		},
		{
			name:    "spanish punctuation",
			input:   "¡hola! ¿qué tal?",
			wantErr: "",
		},
		{
			// Eszett and the oe ligature live above U+00FF so they exercise the
			// multi-byte path through the validator, not just the Latin-1 fast path.
			name:    "eszett and oe ligature",
			input:   "Straße cœur",
			wantErr: "",
		},
		{
			name:    "cjk",
			input:   "中文 日本語 한국어",
			wantErr: "",
		},
		{
			name:    "cyrillic",
			input:   "Привет",
			wantErr: "",
		},
		{
			name:    "greek",
			input:   "Καλημέρα",
			wantErr: "",
		},
		{
			name:    "arabic",
			input:   "مرحبا",
			wantErr: "",
		},
		{
			name:    "hebrew",
			input:   "שלום",
			wantErr: "",
		},
		{
			// If a user wants to put an emoji in their memo we're not going to stop
			// them.
			name:    "emoji",
			input:   "POSTED 💸",
			wantErr: "",
		},
		{
			// Combining marks (the kind that stack on top of a letter) are printable
			// on their own. The 'é' here is actually two runes: 'e' followed by
			// U+0301 combining acute.
			name:    "combining diacritic",
			input:   "é",
			wantErr: "",
		},
		{
			name:    "currency symbols",
			input:   "$ € £ ¥ ₹",
			wantErr: "",
		},
		{
			// The rule itself doesn't care about length. That's [validation.Length]'s
			// job. Pin it just so we don't accidentally start short-circuiting
			// somewhere.
			name:    "long printable input",
			input:   strings.Repeat("á", 1000),
			wantErr: "",
		},

		// Things we want to keep out: control characters.
		{
			name:    "tab",
			input:   "POS\tTED",
			wantErr: wantErr,
		},
		{
			name:    "newline",
			input:   "POS\nTED",
			wantErr: wantErr,
		},
		{
			name:    "carriage return",
			input:   "POS\rTED",
			wantErr: wantErr,
		},
		{
			name:    "nul",
			input:   "POS\x00TED",
			wantErr: wantErr,
		},
		{
			// DEL is the same byte the old [is.PrintableASCII] rule used to catch.
			// Worth pinning here so the swap doesn't accidentally regress that.
			name:    "del",
			input:   "POSTED\x7f",
			wantErr: wantErr,
		},
		{
			name:    "bell",
			input:   "POS\x07TED",
			wantErr: wantErr,
		},
		{
			// The "C1" range from U+0080 to U+009F. ASCII never gets here but a
			// botched encoding conversion can, and we don't want it.
			name:    "c1 control",
			input:   "POSTED",
			wantErr: wantErr,
		},

		// Things we want to keep out: invisible characters that look like regular
		// spaces or look like nothing at all.
		{
			// The classic copy/paste bug. NBSP looks identical to a regular space but
			// compares differently, which causes all kinds of fun downstream.
			name:    "no-break space",
			input:   "hello world",
			wantErr: wantErr,
		},
		{
			name:    "en space",
			input:   "hello world",
			wantErr: wantErr,
		},
		{
			name:    "zero-width joiner",
			input:   "POS‍TED",
			wantErr: wantErr,
		},
		{
			name:    "zero-width non-joiner",
			input:   "POS‌TED",
			wantErr: wantErr,
		},
		{
			// The BOM at the start of a string is the textbook "I copied this from
			// somewhere" giveaway. Spelled with the escape sequence so this file
			// itself stays clean.
			name:    "byte order mark",
			input:   "\uFEFFPOSTED",
			wantErr: wantErr,
		},
		{
			name:    "line separator",
			input:   "POS TED",
			wantErr: wantErr,
		},

		// One bad rune anywhere is enough to fail the whole input. Pin both ends so
		// we don't accidentally start short-circuiting.
		{
			name:    "valid prefix then control char",
			input:   "Pösted\t",
			wantErr: wantErr,
		},
		{
			name:    "control char then valid suffix",
			input:   "\tPösted",
			wantErr: wantErr,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validators.PrintableUnicode.Validate(tc.input)
			if tc.wantErr == "" {
				assert.NoError(t, err, "input must be accepted")
			} else {
				assert.EqualError(t, err, tc.wantErr, "input must be rejected with the expected message")
			}
		})
	}
}

func TestPrintableUnicode_ErrorIdentity(t *testing.T) {
	// The error the rule reports comes from [validators.ErrPrintableUnicode] so
	// callers can match on the sentinel rather than scraping the message text.
	// The "validation_is_*" code lines up with the other rules in the validation
	// library, which gives us a consistent shape on the wire.
	err := validators.PrintableUnicode.Validate("bad\tinput")
	assert.Error(t, err, "control character must surface an error")

	verr, ok := err.(interface {
		Code() string
		Message() string
	})
	if assert.True(t, ok, "rule error must implement validation.Error") {
		assert.Equal(t, "validation_is_printable_unicode", verr.Code(), "error code must be stable for clients")
		assert.Equal(t, "must contain printable characters only", verr.Message(), "default message must be stable")
	}
}
