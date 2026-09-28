package currency

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupplementalCurrencyParsing(t *testing.T) {
	currencyData, err := cldrDataset.Open("sources/currencyData.json")
	if err != nil {
		fmt.Printf("Failed to load CLDR supplemental currency data: %+v\n", err)
	}

	var data supplementalCurrencyData
	err = json.NewDecoder(currencyData).Decode(&data)
	assert.NoError(t, err)
}

func TestGetCurrencyWithoutCLDRNames(t *testing.T) {
	t.Run("falls back to the code", func(t *testing.T) {
		// Every real currency has a name in en, so fake one that CLDR knows
		// nothing about
		fractionalDigits["ZZZ"] = 2
		t.Cleanup(func() {
			delete(fractionalDigits, "ZZZ")
		})

		result, err := GetCurrency("en", "ZZZ")
		assert.NoError(t, err)
		assert.Equal(t, "ZZZ", result.Name)
		assert.Equal(t, "ZZZ", result.Symbol)
		assert.EqualValues(t, 2, result.FractionalDigits)
	})
}
