// cldrgen takes a checkout of cldr-json and squashes the per-locale currency
// data down into one file that the currency package embeds. The raw data is
// like 50MB and most of it is either currencies we don't support or the same
// value as the parent locale, so this only keeps what actually differs
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
				// A handful of currencies use their own separators, like CVE in pt-CV
				// is written as 1$50
				Decimal string `json:"decimal"`
				Group   string `json:"group"`
			} `json:"currencies"`
		} `json:"numbers"`
	} `json:"main"`
}

type localeNumbers struct {
	Main map[string]struct {
		Numbers struct {
			// Some locales default to other numbering systems like Arabic-Indic
			// digits, but monetr only parses ASCII digits so always use the latin
			// symbols
			Symbols struct {
				Decimal string `json:"decimal"`
				Group   string `json:"group"`
			} `json:"symbols-numberSystem-latn"`
		} `json:"numbers"`
	} `json:"main"`
}

type parentLocales struct {
	Supplemental struct {
		ParentLocales struct {
			ParentLocale map[string]string `json:"parentLocale"`
		} `json:"parentLocales"`
	} `json:"supplemental"`
}

type likelySubtags struct {
	Supplemental struct {
		LikelySubtags map[string]string `json:"likelySubtags"`
	} `json:"supplemental"`
}

// Output format, keyed by locale. Each locale has its CLDR parent, its decimal
// and grouping separators and then its currencies keyed by code. Anything
// blank gets inherited from the parent locale
type currencyLocale struct {
	Parent     string                        `json:"p,omitempty"`
	Decimal    string                        `json:"d,omitempty"`
	Group      string                        `json:"g,omitempty"`
	Currencies map[string]currencyLocaleData `json:"c,omitempty"`
}

type currencyLocaleData struct {
	Name    string `json:"n,omitempty"`
	Symbol  string `json:"s,omitempty"`
	Decimal string `json:"d,omitempty"`
	Group   string `json:"g,omitempty"`
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
	// comes up with at runtime
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

	var parents parentLocales
	if err := readJSON(
		filepath.Join(cldrPath, "cldr-json/cldr-core/supplemental/parentLocales.json"),
		&parents,
	); err != nil {
		return err
	}
	var likely likelySubtags
	if err := readJSON(
		filepath.Join(cldrPath, "cldr-json/cldr-core/supplemental/likelySubtags.json"),
		&likely,
	); err != nil {
		return err
	}

	mainPath := filepath.Join(cldrPath, "cldr-json/cldr-numbers-full/main")
	dirs, err := os.ReadDir(mainPath)
	if err != nil {
		return err
	}

	raw := map[string]map[string]currencyLocaleData{}
	rawSymbols := map[string]currencyLocale{}
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
				Name:    currency.DisplayName,
				Symbol:  currency.Symbol,
				Decimal: currency.Decimal,
				Group:   currency.Group,
			}
		}
		raw[locale] = values

		var numbers localeNumbers
		if err := readJSON(filepath.Join(mainPath, locale, "numbers.json"), &numbers); err != nil {
			return err
		}
		rawSymbols[locale] = currencyLocale{
			Decimal: numbers.Main[locale].Numbers.Symbols.Decimal,
			Group:   numbers.Main[locale].Numbers.Symbols.Group,
		}
	}

	result := map[string]currencyLocale{}
	for locale, values := range raw {
		parent := parentOf(
			raw,
			parents.Supplemental.ParentLocales.ParentLocale,
			likely.Supplemental.LikelySubtags,
			locale,
		)
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
			if value.Decimal != inherited.Decimal {
				item.Decimal = value.Decimal
			}
			if value.Group != inherited.Group {
				item.Group = value.Group
			}
			if item != (currencyLocaleData{}) {
				diff[code] = item
			}
		}
		// Always include the locale even if nothing is different, that way the
		// currency package still knows the locale exists and what its parent is
		item := currencyLocale{
			Parent:     parent,
			Currencies: diff,
		}
		var inherited currencyLocale
		if parent != "" {
			inherited = rawSymbols[parent]
		}
		if symbols := rawSymbols[locale]; symbols.Decimal != inherited.Decimal {
			item.Decimal = symbols.Decimal
		}
		if symbols := rawSymbols[locale]; symbols.Group != inherited.Group {
			item.Group = symbols.Group
		}
		result[locale] = item
	}

	output, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, output, 0o644)
}

// parentOf returns the parent of the locale using CLDR's inheritance rules and
// keeps walking up until it finds a parent that actually exists in the dataset.
// Returns blank for root (und) since thats the top. The rules are:
//   - If CLDR has an explicit parent for the locale then use that, this is how
//     things like es-MX get es-419 instead of es
//   - If the locale is just a language and a script, and that script isn't the
//     likely script for that language then the parent is root. Otherwise
//     az-Cyrl would inherit latin script names from az
//   - Otherwise just chop off the last subtag, plain languages go to root
//
// https://www.unicode.org/reports/tr35/#Parent_Locales
func parentOf(
	locales map[string]map[string]currencyLocaleData,
	parents map[string]string,
	likely map[string]string,
	locale string,
) string {
	for locale != "und" {
		if parent, ok := parents[locale]; ok {
			locale = parent
		} else if parts := strings.Split(locale, "-"); len(parts) == 2 && len(parts[1]) == 4 &&
			!strings.Contains(likely[parts[0]], "-"+parts[1]+"-") {
			locale = "und"
		} else if i := strings.LastIndex(locale, "-"); i >= 0 {
			locale = locale[:i]
		} else {
			locale = "und"
		}

		if _, ok := locales[locale]; ok {
			return locale
		}
	}
	return ""
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
