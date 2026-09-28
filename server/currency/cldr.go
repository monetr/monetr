package currency

import (
	"embed"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

//go:embed sources/currencyData.json sources/likelySubtags.json sources/currencyNames.json
var cldrDataset embed.FS

type supplementalVersion struct {
	UnicodeVersion string `json:"_unicodeVersion"`
	CLDRVersion    string `json:"_cldrVersion"`
}

type supplementalCurrencyData struct {
	Supplemental struct {
		Version      supplementalVersion `json:"version"`
		CurrencyData struct {
			Fractions map[string]supplementalCurrencyFraction            `json:"fractions"`
			Region    map[string][]map[string]supplementalCurrencyRegion `json:"region"`
		} `json:"currencyData"`
	} `json:"supplemental"`
}

type supplementalCurrencyFraction struct {
	Rounding     string `json:"_rounding"`
	Digits       string `json:"_digits"`
	CashRounding string `json:"_cashRounding"`
	CashDigits   string `json:"_cashDigits"`
}

type supplementalCurrencyRegion struct {
	From   *jsonDate `json:"_from"`
	To     *jsonDate `json:"_to"`
	Tender *string   `json:"_tender"`
}

// isCurrent returns true if the currency is still used in the region and is
// actually legal tender
func (s supplementalCurrencyRegion) isCurrent() bool {
	return s.To == nil && (s.Tender == nil || *s.Tender != "false")
}

type supplementalLikelySubtags struct {
	Supplemental struct {
		LikelySubtags map[string]string `json:"likelySubtags"`
	} `json:"supplemental"`
}

// This is the format of the currencyNames.json file that cldrgen spits out,
// anything blank gets inherited from the parent locale
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

type jsonDate time.Time

func (j *jsonDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*j).Format(time.DateOnly))
}

func (j *jsonDate) UnmarshalJSON(input []byte) error {
	inputStr := string(input)
	// Need to remove leading and trailing double quotes too.
	inputStr = strings.Trim(inputStr, `"`)
	result, err := time.Parse(time.DateOnly, inputStr)
	if err != nil {
		return err
	}
	*j = jsonDate(result)
	return nil
}

var (
	// All of this gets populated once from the embedded CLDR dataset in init and
	// is only read after that
	cldrVersion      string
	currencies       []string
	fractionalDigits map[string]int64
	regionCurrencies map[string]string
	likelySubtags    map[string]string
	currencyNames    map[string]currencyLocale
	// Needs to be in the same order as the tags given to localeMatcher
	locales       []string
	localeMatcher language.Matcher
)

// defaultLocale is what we use when the requested locale doesn't match anything
// we have CLDR data for
const defaultLocale = "en"

func init() {
	if err := loadCLDR(); err != nil {
		panic(fmt.Sprintf("failed to load embedded CLDR dataset: %+v", err))
	}
}

func loadCLDR() error {
	var currencyData supplementalCurrencyData
	if err := readCLDRFile("sources/currencyData.json", &currencyData); err != nil {
		return err
	}
	var subtags supplementalLikelySubtags
	if err := readCLDRFile("sources/likelySubtags.json", &subtags); err != nil {
		return err
	}

	data := currencyData.Supplemental.CurrencyData
	// Anything not in the fractions map uses the DEFAULT entry
	defaultDigits, err := strconv.ParseInt(data.Fractions["DEFAULT"].Digits, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse default currency digits")
	}

	fractionalDigits = map[string]int64{}
	regionCurrencies = map[string]string{}
	for region, entries := range data.Region {
		for _, entry := range entries {
			for code, details := range entry {
				if !details.isCurrent() {
					continue
				}

				// Regions list their current currencies first, so the first current one
				// we see is the primary currency for that region
				if _, ok := regionCurrencies[region]; !ok {
					regionCurrencies[region] = code
				}

				if _, ok := fractionalDigits[code]; ok {
					continue
				}
				digits := defaultDigits
				if fraction, ok := data.Fractions[code]; ok {
					digits, err = strconv.ParseInt(fraction.Digits, 10, 64)
					if err != nil {
						return errors.Wrapf(err, "failed to parse digits for currency [%s]", code)
					}
				}
				fractionalDigits[code] = digits
			}
		}
	}

	currencies = make([]string, 0, len(fractionalDigits))
	for code := range fractionalDigits {
		currencies = append(currencies, code)
	}
	slices.Sort(currencies)

	if err := readCLDRFile("sources/currencyNames.json", &currencyNames); err != nil {
		return err
	}
	if _, ok := currencyNames[defaultLocale]; !ok {
		return errors.Errorf("default locale [%s] is missing from currency names", defaultLocale)
	}
	// The matcher falls back to the first tag when nothing else matches, so the
	// default locale has to be first
	locales = []string{defaultLocale}
	for locale := range currencyNames {
		if locale != defaultLocale {
			locales = append(locales, locale)
		}
	}
	slices.Sort(locales[1:])
	tags := make([]language.Tag, 0, len(locales))
	for _, locale := range locales {
		tag, err := language.Parse(locale)
		if err != nil {
			return errors.Wrapf(err, "failed to parse CLDR locale [%s]", locale)
		}
		tags = append(tags, tag)
	}
	localeMatcher = language.NewMatcher(tags)

	cldrVersion = currencyData.Supplemental.Version.CLDRVersion
	likelySubtags = subtags.Supplemental.LikelySubtags
	return nil
}

func readCLDRFile(path string, result any) error {
	file, err := cldrDataset.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open CLDR file [%s]", path)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(result); err != nil {
		return errors.Wrapf(err, "failed to parse CLDR file [%s]", path)
	}
	return nil
}

// GetCLDRVersion returns the version of the CLDR dataset embedded in monetr
func GetCLDRVersion() string {
	return cldrVersion
}

// GetCurrencies returns a sorted list of all the ISO 4217 currency codes that
// monetr supports, its a copy so its fine to modify it
func GetCurrencies() []string {
	return slices.Clone(currencies)
}

// IsSupportedCurrency returns true if the provided currency code is one that
// monetr supports
func IsSupportedCurrency(currency string) bool {
	_, ok := fractionalDigits[currency]
	return ok
}

// GetFractionalDigits returns the number of digits after the decimal place
// that the provided currency uses, like USD uses 2 and JPY uses 0
func GetFractionalDigits(currency string) (int64, error) {
	digits, ok := fractionalDigits[currency]
	if !ok {
		return 0, errors.New("currency not supported")
	}
	return digits, nil
}

// GetCurrencyForLocale returns the currency that is currently legal tender in
// the region for the provided locale. The locale can be BCP 47 like "en-US" or
// POSIX like "en_US.UTF-8", if there isn't a region in the locale then we use
// the most likely region for the language. So "ja" will resolve to JPY
func GetCurrencyForLocale(locale string) (string, error) {
	region, err := getRegionForLocale(locale)
	if err != nil {
		return "", err
	}
	code, ok := regionCurrencies[region]
	if !ok {
		return "", errors.Errorf("no currency found for locale [%s] region [%s]", locale, region)
	}
	return code, nil
}

// ParseLocale takes a locale in either BCP 47 form like "en-US" or POSIX form
// like "en_US.UTF-8" or "de_DE@euro" and parses it into a language tag.
// Anything that takes a locale from the outside world should go through this
// first
func ParseLocale(locale string) (language.Tag, error) {
	// Strip off any POSIX encoding or modifier, the tag doesn't care about them
	if i := strings.IndexAny(locale, ".@"); i >= 0 {
		locale = locale[:i]
	}
	locale = strings.ReplaceAll(strings.TrimSpace(locale), "_", "-")
	if locale == "" {
		return language.Und, errors.New("locale cannot be blank")
	}
	tag, err := language.Parse(locale)
	if err != nil {
		return language.Und, errors.Wrapf(err, "failed to parse locale [%s]", locale)
	}
	return tag, nil
}

func getRegionForLocale(locale string) (string, error) {
	tag, err := ParseLocale(locale)
	if err != nil {
		return "", err
	}

	// If the locale actually has a region in it then just use that
	if region, confidence := tag.Region(); confidence == language.Exact {
		return region.String(), nil
	}

	// Otherwise use CLDR's likely subtags to figure out the region. Try the
	// language and script together first since that can change the answer, zh
	// is China but zh-Hant is Taiwan
	base, _ := tag.Base()
	keys := []string{base.String()}
	if script, confidence := tag.Script(); confidence == language.Exact {
		keys = append([]string{base.String() + "-" + script.String()}, keys...)
	}
	for _, key := range keys {
		likely, ok := likelySubtags[key]
		if !ok {
			continue
		}
		likelyTag, err := language.Parse(likely)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse likely subtag [%s]", likely)
		}
		if region, confidence := likelyTag.Region(); confidence == language.Exact {
			return region.String(), nil
		}
	}
	return "", errors.Errorf("could not determine region for locale [%s]", locale)
}

// MatchLocale takes the value of an Accept-Language header and returns the
// closest locale that monetr has currency names for. If nothing matches or the
// header is blank or garbage then you get the default locale (en)
func MatchLocale(acceptLanguage string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return defaultLocale
	}
	_, index, _ := localeMatcher.Match(tags...)
	return locales[index]
}

// GetCurrency returns the details of a single currency with its name and
// symbol localized for the provided locale, the locale should be one that came
// from [MatchLocale]
func GetCurrency(locale, code string) (Currency, error) {
	digits, err := GetFractionalDigits(code)
	if err != nil {
		return Currency{}, err
	}

	result := Currency{
		Code:             code,
		FractionalDigits: digits,
	}
	// Walk up from the requested locale through its parents and then the
	// default locale, taking the first value we find for each field
	for _, candidate := range append(localeChain(locale), defaultLocale) {
		data := currencyNames[candidate].Currencies[code]
		if result.Name == "" {
			result.Name = data.Name
		}
		if result.Symbol == "" {
			result.Symbol = data.Symbol
		}
		if result.Name != "" && result.Symbol != "" {
			break
		}
	}
	// If CLDR has nothing at all then the code is the best we can do
	if result.Name == "" {
		result.Name = code
	}
	if result.Symbol == "" {
		result.Symbol = code
	}
	return result, nil
}

// GetSeparators returns the decimal and grouping separators used when writing
// an amount of the provided currency in the provided locale. Like "." and ","
// for USD in en or "," and "." for EUR in de. These are always the separators
// for latin (ASCII) digits, and the locale should be one that came from
// [MatchLocale]. A few currencies have their own separators in some locales
// (CVE in pt-CV is written as 1$50) so those win over the locale's separators
func GetSeparators(locale, code string) (decimal, group string) {
	chain := append(localeChain(locale), defaultLocale)
	// Look for currency specific separators first
	for _, candidate := range chain {
		data := currencyNames[candidate].Currencies[code]
		if decimal == "" {
			decimal = data.Decimal
		}
		if group == "" {
			group = data.Group
		}
	}
	// Then fill in whatever is left from the locale itself
	for _, candidate := range chain {
		data := currencyNames[candidate]
		if decimal == "" {
			decimal = data.Decimal
		}
		if group == "" {
			group = data.Group
		}
	}
	// Root always has these so this shouldn't happen, but just in case
	if decimal == "" {
		decimal = "."
	}
	if group == "" {
		group = ","
	}
	return decimal, group
}

// GetCurrencyList returns every supported currency sorted by code, with the
// names and symbols localized for the provided locale
func GetCurrencyList(locale string) []Currency {
	result := make([]Currency, 0, len(currencies))
	for _, code := range currencies {
		// Every code in currencies is supported so this can't fail
		item, _ := GetCurrency(locale, code)
		result = append(result, item)
	}
	return result
}

// localeChain returns the locale followed by each of its parents. cldrgen
// figures out the parents using CLDR's inheritance rules, so zh-Hant-HK
// returns zh-Hant-HK, zh-Hant, und
func localeChain(locale string) []string {
	chain := []string{}
	for locale != "" {
		// Should never happen, but don't spin forever if the data is bad
		if len(chain) > len(currencyNames) {
			break
		}
		chain = append(chain, locale)
		locale = currencyNames[locale].Parent
	}
	return chain
}
