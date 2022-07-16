//go:build icons && simple_icons

package icons

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed sources/simple-icons/package.json
//go:embed sources/simple-icons/slugs.md
//go:embed sources/simple-icons/_data/simple-icons.json
//go:embed sources/simple-icons/icons/*.svg
var simpleIconsFiles embed.FS

var (
	_ IconIndex = &simpleIconsIndex{}
)

type simpleIconsIndex struct {
	slugs    map[string]Icon
	searches map[string]string
	version  string
}

func (s *simpleIconsIndex) Search(input string) *Icon {
	bySlug, ok := s.slugs[strings.ToLower(input)]
	if ok {
		return &bySlug
	}

	// TODO make all searches lowercase
	slug, ok := s.searches[input]
	if ok {
		bySlug = s.slugs[slug]
		return &bySlug
	}

	return nil
}

func (s simpleIconsIndex) Name() string {
	if s.version != "" {
		return fmt.Sprintf("simple-icons@%s", s.version)
	}

	return "simple-icons"
}

func newSimpleIconsIndex() *simpleIconsIndex {
	slugFile, err := simpleIconsFiles.ReadFile("sources/simple-icons/slugs.md")
	if err != nil {
		// Could not initialize simple icons
		return nil
	}
	nameToSlug, err := parseSimpleIconsSlugMarkdown(slugFile)
	if err != nil {
		return nil
	}

	type Metadata struct {
		Title string `json:"title"`
		Hex   string `json:"hex"`
	}

	var metadata struct {
		Icons []Metadata `json:"icons"`
	}
	metadataBytes, _ := simpleIconsFiles.ReadFile("sources/simple-icons/_data/simple-icons.json")
	if err == nil {
		_ = json.Unmarshal(metadataBytes, &metadata)
	}

	icons := map[string]Icon{}
	for title, slug := range nameToSlug {
		iconFile, err := simpleIconsFiles.ReadFile(fmt.Sprintf("sources/simple-icons/icons/%s.svg", slug))
		if err != nil {
			return nil
		}

		dereferenceTitle := title
		data := Icon{
			Title:   &dereferenceTitle,
			Slug:    slug,
			Library: "simple-icons",
			SVG:     base64.StdEncoding.EncodeToString(iconFile),
			Colors:  nil,
		}
		iconMetadata := func(name string) *Metadata {
			for _, item := range metadata.Icons {
				if strings.EqualFold(item.Title, name) {
					return &item
				}
			}

			return nil
		}(title)
		if iconMetadata != nil {
			data.Colors = []string{
				iconMetadata.Hex,
			}
		}

		icons[slug] = data
	}

	var packageInfo struct {
		Version string `json:"version"`
	}
	packageJsonBytes, _ := simpleIconsFiles.ReadFile("sources/simple-icons/package.json")
	if err == nil {
		_ = json.Unmarshal(packageJsonBytes, &packageInfo)
	}


	return &simpleIconsIndex{
		slugs:    icons,
		searches: nameToSlug,
		version: packageInfo.Version,
	}
}

func init() {
	simpleIcons := newSimpleIconsIndex()
	if simpleIcons == nil {
		return
	}

	indexes = append(indexes, simpleIcons)
}
