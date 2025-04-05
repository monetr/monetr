package currency

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupplementalCurrencyParsing(t *testing.T) {
	currencyData, err := cldrDataset.Open("sources/cldr-core/supplemental/currencyData.json")
	if err != nil {
		fmt.Printf("Failed to load CLDR supplemental currency data: %+v\n", err)
	}

	var data supplementalCurrencyData
	err = json.NewDecoder(currencyData).Decode(&data)
	assert.NoError(t, err)
}
