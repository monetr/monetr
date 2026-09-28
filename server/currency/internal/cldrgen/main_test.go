package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParentOf(t *testing.T) {
	// Just enough of the dataset to hit each of the rules
	locales := map[string]map[string]currencyLocaleData{
		"und":     {},
		"en":      {},
		"en-001":  {},
		"en-AU":   {},
		"es":      {},
		"es-419":  {},
		"es-MX":   {},
		"de":      {},
		"de-CH":   {},
		"az":      {},
		"az-Cyrl": {},
		"az-Latn": {},
	}
	parents := map[string]string{
		"en-AU":  "en-001",
		"es-MX":  "es-419",
		"en-001": "en",
	}
	likely := map[string]string{
		"az": "az-Latn-AZ",
		"de": "de-Latn-DE",
		"en": "en-Latn-US",
	}

	t.Run("explicit parent", func(t *testing.T) {
		assert.Equal(t, "es-419", parentOf(locales, parents, likely, "es-MX"))
		assert.Equal(t, "en-001", parentOf(locales, parents, likely, "en-AU"))
	})

	t.Run("truncation", func(t *testing.T) {
		assert.Equal(t, "de", parentOf(locales, parents, likely, "de-CH"))
		// es-419 has no explicit parent so it just gets chopped down to es
		assert.Equal(t, "es", parentOf(locales, parents, likely, "es-419"))
	})

	t.Run("unlikely script goes to root", func(t *testing.T) {
		assert.Equal(t, "und", parentOf(locales, parents, likely, "az-Cyrl"))
	})

	t.Run("likely script goes to the language", func(t *testing.T) {
		assert.Equal(t, "az", parentOf(locales, parents, likely, "az-Latn"))
	})

	t.Run("language goes to root", func(t *testing.T) {
		assert.Equal(t, "und", parentOf(locales, parents, likely, "de"))
	})

	t.Run("skips parents that are not in the dataset", func(t *testing.T) {
		// de-CH-1996 isn't in the dataset but de-CH is, and de-AT isn't in the
		// dataset either so that one should skip right to de
		assert.Equal(t, "de-CH", parentOf(locales, parents, likely, "de-CH-1996"))
		assert.Equal(t, "de", parentOf(locales, parents, likely, "de-AT-1996"))
	})

	t.Run("root has no parent", func(t *testing.T) {
		assert.Empty(t, parentOf(locales, parents, likely, "und"))
	})
}
