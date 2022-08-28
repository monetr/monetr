package icons

import (
	"strings"
)

type IconRepository struct {
}

type Icon struct {
	Title   *string  `json:"title"`
	Slug    string   `json:"slug"`
	Library string   `json:"library"`
	SVG     string   `json:"svg"`
	Colors  []string `json:"colors"`
}

func SearchIcon(input string) (*Icon, error) {
	// Search for icon using the parts of the input string as well as the string with no spaces. This helps account for
	// some minor mismatches in naming between our icon slugs and transaction/merchant names.
	parts := append(strings.Split(input, " "), strings.ReplaceAll(input, " ", ""))
	for _, part := range parts {
		for _, index := range indexes {
			if icon := index.Search(part); icon != nil {
				return icon, nil
			}
		}
	}

	return nil, nil
}
