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
	parts := strings.Split(input, " ")
	for _, part := range parts {
		for _, index := range indexes {
			if icon := index.Search(part); icon != nil {
				return icon, nil
			}
		}
	}

	return nil, nil
}
