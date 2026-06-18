package currency

import (
	"embed"
	"encoding/json"
	"strings"
	"time"
)

//go:embed sources/cldr-core/supplemental/currencyData.json sources/cldr-numbers-full/main/*
var cldrDataset embed.FS

type supplementalCurrencyData struct {
	Supplemental struct {
		CurrencyData struct {
			Fractions map[string]struct {
				Rounding     string `json:"_rounding"`
				Digits       string `json:"_digits"`
				CashRounding string `json:"_cashRounding"`
				CashDigits   string `json:"_cashDigits"`
			} `json:"fractions"`
			Region map[string][]map[string]supplementalCurrencyRegion
		} `json:"currencyData"`
	} `json:"supplemental"`
}

type supplementalCurrencyRegion struct {
	From   *jsonDate `json:"_from"`
	To     *jsonDate `json:"_to"`
	Tender *string   `json:"_tender"`
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

func init() {

}
