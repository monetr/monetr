//go:build icons && simple_icons

package icons

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed sources/simple-icons/LICENSE.md
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

	var metadata []Metadata
	metadataBytes, err := simpleIconsFiles.ReadFile("sources/simple-icons/_data/simple-icons.json")
	if err == nil {
		_ = json.Unmarshal(metadataBytes, &metadata)
	}

	// Index the colors by lowercase title up front, scanning the metadata for
	// every icon is quadratic and was the bulk of the cost of building this
	// index. Keep the first entry for a title to match the previous scan.
	colorsByTitle := make(map[string]string, len(metadata))
	for _, item := range metadata {
		title := strings.ToLower(item.Title)
		if _, ok := colorsByTitle[title]; !ok {
			colorsByTitle[title] = item.Hex
		}
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
		if hex, ok := colorsByTitle[strings.ToLower(title)]; ok {
			data.Colors = []string{
				hex,
			}
		}

		icons[slug] = data
	}

	var packageInfo struct {
		Version string `json:"version"`
	}
	packageJsonBytes, err := simpleIconsFiles.ReadFile("sources/simple-icons/package.json")
	if err == nil {
		_ = json.Unmarshal(packageJsonBytes, &packageInfo)
	}

	return &simpleIconsIndex{
		slugs:    icons,
		searches: nameToSlug,
		version:  packageInfo.Version,
	}
}

func init() {
	registerIndex(func() IconIndex {
		simpleIcons := newSimpleIconsIndex()
		// Return an untyped nil, a nil *simpleIconsIndex would be a non-nil
		// IconIndex.
		if simpleIcons == nil {
			return nil
		}

		return simpleIcons
	})
}
