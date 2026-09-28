// cldrgen reads the per-locale currency names and symbols out of a checkout of
// the cldr-json repository and writes a single compact JSON file that the
// currency package embeds. The raw CLDR data is roughly 50MB across hundreds of
// locales, most of which is either currencies monetr does not support or values
// identical to the parent locale. So this only keeps the supported currencies,
// and only keeps a locale's value when it differs from its parent's.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type currencyData struct {
	Supplemental struct {
		CurrencyData struct {
			Region map[string][]map[string]struct {
				To     *string `json:"_to"`
				Tender *string `json:"_tender"`
			} `json:"region"`
		} `json:"currencyData"`
	} `json:"supplemental"`
}

type localeCurrencies struct {
	Main map[string]struct {
		Numbers struct {
			Currencies map[string]struct {
				DisplayName string `json:"displayName"`
				Symbol      string `json:"symbol"`
			} `json:"currencies"`
		} `json:"numbers"`
	} `json:"main"`
}

// Output format, keyed by locale then by currency code. Blank fields are
// inherited from the parent locale.
type currencyLocaleData struct {
	Name   string `json:"n,omitempty"`
	Symbol string `json:"s,omitempty"`
}

func main() {
	cldrPath := flag.String("cldr", "", "path to the root of the cldr-json checkout")
	outputPath := flag.String("output", "", "path to write the generated json file to")
	flag.Parse()
	if *cldrPath == "" || *outputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*cldrPath, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate currency names: %+v\n", err)
		os.Exit(1)
	}
}

func run(cldrPath, outputPath string) error {
	var supplemental currencyData
	if err := readJSON(
		filepath.Join(cldrPath, "cldr-json/cldr-core/supplemental/currencyData.json"),
		&supplemental,
	); err != nil {
		return err
	}

	// Build the same set of supported currencies that the currency package
	// derives at runtime.
	supported := map[string]bool{}
	for _, entries := range supplemental.Supplemental.CurrencyData.Region {
		for _, entry := range entries {
			for code, details := range entry {
				if details.To == nil && (details.Tender == nil || *details.Tender != "false") {
					supported[code] = true
				}
			}
		}
	}

	mainPath := filepath.Join(cldrPath, "cldr-json/cldr-numbers-full/main")
	dirs, err := os.ReadDir(mainPath)
	if err != nil {
		return err
	}

	raw := map[string]map[string]currencyLocaleData{}
	for _, dir := range dirs {
		locale := dir.Name()
		var data localeCurrencies
		if err := readJSON(filepath.Join(mainPath, locale, "currencies.json"), &data); err != nil {
			return err
		}
		values := map[string]currencyLocaleData{}
		for code, currency := range data.Main[locale].Numbers.Currencies {
			if !supported[code] {
				continue
			}
			values[code] = currencyLocaleData{
				Name:   currency.DisplayName,
				Symbol: currency.Symbol,
			}
		}
		raw[locale] = values
	}

	result := map[string]map[string]currencyLocaleData{}
	for locale, values := range raw {
		parent := parentOf(raw, locale)
		diff := map[string]currencyLocaleData{}
		for code, value := range values {
			var inherited currencyLocaleData
			if parent != "" {
				inherited = raw[parent][code]
			}
			item := currencyLocaleData{}
			if value.Name != inherited.Name {
				item.Name = value.Name
			}
			if value.Symbol != inherited.Symbol {
				item.Symbol = value.Symbol
			}
			if item != (currencyLocaleData{}) {
				diff[code] = item
			}
		}
		// Always include the locale, even when it has no differences, so that the
		// currency package knows the locale exists.
		result[locale] = diff
	}

	output, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, output, 0o644)
}

// parentOf returns the closest locale that exists in the dataset by truncating
// the locale one subtag at a time. So zh-Hant-TW becomes zh-Hant and then zh.
// This must match the fallback used by the currency package.
func parentOf(locales map[string]map[string]currencyLocaleData, locale string) string {
	for {
		i := strings.LastIndex(locale, "-")
		if i < 0 {
			return ""
		}
		locale = locale[:i]
		if _, ok := locales[locale]; ok {
			return locale
		}
	}
}

func readJSON(path string, result any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(result); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return nil
}
