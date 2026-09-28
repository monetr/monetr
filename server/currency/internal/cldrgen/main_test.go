package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRun(t *testing.T) {
	writeJSON := func(t *testing.T, path string, value any) {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		data, err := json.Marshal(value)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(path, data, 0o644))
	}
	writeLocale := func(
		t *testing.T,
		root, locale string,
		currencies map[string]any,
		symbols map[string]string,
	) {
		dir := filepath.Join(root, "cldr-json/cldr-numbers-full/main", locale)
		writeJSON(t, filepath.Join(dir, "currencies.json"), map[string]any{
			"main": map[string]any{
				locale: map[string]any{
					"numbers": map[string]any{
						"currencies": currencies,
					},
				},
			},
		})
		writeJSON(t, filepath.Join(dir, "numbers.json"), map[string]any{
			"main": map[string]any{
				locale: map[string]any{
					"numbers": map[string]any{
						"symbols-numberSystem-latn": symbols,
					},
				},
			},
		})
	}

	// Just enough of a cldr-json checkout to hit each of the rules
	root := t.TempDir()
	supplemental := filepath.Join(root, "cldr-json/cldr-core/supplemental")
	writeJSON(t, filepath.Join(supplemental, "currencyData.json"), map[string]any{
		"supplemental": map[string]any{
			"currencyData": map[string]any{
				"region": map[string]any{
					"US": []map[string]any{
						{"USD": map[string]any{"_from": "1792-01-01"}},
						{"USN": map[string]any{"_tender": "false"}},
					},
					"AU": []map[string]any{
						{"AUD": map[string]any{"_from": "1966-02-14"}},
					},
					"DE": []map[string]any{
						{"DEM": map[string]any{"_from": "1948-06-20", "_to": "2002-02-28"}},
					},
				},
			},
		},
	})
	writeJSON(t, filepath.Join(supplemental, "parentLocales.json"), map[string]any{
		"supplemental": map[string]any{
			"parentLocales": map[string]any{
				"parentLocale": map[string]string{},
			},
		},
	})
	writeJSON(t, filepath.Join(supplemental, "likelySubtags.json"), map[string]any{
		"supplemental": map[string]any{
			"likelySubtags": map[string]string{
				"en": "en-Latn-US",
			},
		},
	})

	rootSymbols := map[string]string{"decimal": ".", "group": ",", "minusSign": "-"}
	writeLocale(t, root, "und", map[string]any{
		"USD": map[string]string{"displayName": "USD", "symbol": "US$"},
	}, rootSymbols)
	writeLocale(t, root, "en", map[string]any{
		"USD": map[string]string{"displayName": "US Dollar", "symbol": "$"},
		"AUD": map[string]string{"displayName": "Australian Dollar", "symbol": "A$"},
		"USN": map[string]string{"displayName": "US Dollar (Next day)", "symbol": "USN"},
		"DEM": map[string]string{"displayName": "German Mark", "symbol": "DEM"},
	}, rootSymbols)
	writeLocale(t, root, "en-AU", map[string]any{
		// Same name as en but a different symbol
		"USD": map[string]string{"displayName": "US Dollar", "symbol": "USD"},
		// Exactly the same as en
		"AUD": map[string]string{"displayName": "Australian Dollar", "symbol": "A$"},
	}, rootSymbols)
	writeLocale(t, root, "en-DE", map[string]any{}, map[string]string{
		"decimal":   ",",
		"group":     ".",
		"minusSign": "-",
	})

	output := filepath.Join(t.TempDir(), "currencyNames.json")
	require.NoError(t, run(root, output))

	data, err := os.ReadFile(output)
	require.NoError(t, err)
	var result map[string]currencyLocale
	require.NoError(t, json.Unmarshal(data, &result))

	t.Run("root keeps everything", func(t *testing.T) {
		assert.Equal(t, currencyLocale{
			Decimal: ".",
			Group:   ",",
			Minus:   "-",
			Currencies: map[string]currencyLocaleData{
				"USD": {Name: "USD", Symbol: "US$"},
			},
		}, result["und"])
	})

	t.Run("only supported currencies are kept", func(t *testing.T) {
		// USN is not tender and DEM is not current anymore
		assert.Equal(t, currencyLocale{
			Parent: "und",
			Currencies: map[string]currencyLocaleData{
				"USD": {Name: "US Dollar", Symbol: "$"},
				"AUD": {Name: "Australian Dollar", Symbol: "A$"},
			},
		}, result["en"])
	})

	t.Run("only differences from the parent are kept", func(t *testing.T) {
		assert.Equal(t, currencyLocale{
			Parent: "en",
			Currencies: map[string]currencyLocaleData{
				"USD": {Symbol: "USD"},
			},
		}, result["en-AU"])
	})

	t.Run("locale symbols that differ from the parent are kept", func(t *testing.T) {
		assert.Equal(t, currencyLocale{
			Parent:  "en",
			Decimal: ",",
			Group:   ".",
		}, result["en-DE"])
	})
}

func TestRunMissingData(t *testing.T) {
	t.Run("missing checkout", func(t *testing.T) {
		err := run(filepath.Join(t.TempDir(), "nope"), filepath.Join(t.TempDir(), "out.json"))
		assert.Error(t, err)
	})

	t.Run("bad json", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "cldr-json/cldr-core/supplemental/currencyData.json")
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte("{"), 0o644))
		err := run(root, filepath.Join(t.TempDir(), "out.json"))
		assert.ErrorContains(t, err, "failed to parse")
	})
}
