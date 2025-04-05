package currency

import (
	"embed"
	"fmt"
)

//go:embed sources/cldr-core/supplemental/currencyData.json sources/cldr-numbers-full/main/*
var cldrDataset embed.FS

func init() {
	fmt.Sprint(cldrDataset.Open)
}
