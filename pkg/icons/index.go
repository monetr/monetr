package icons

type IconIndex interface {
	Search(input string) *Icon
	Name() string
}

var (
	enabled bool
	indexes []IconIndex
)

func GetIconsEnabled() bool {
	return enabled
}

func GetIconIndexes() []string {
	names := make([]string, len(indexes))
	for i, index := range indexes {
		names[i] = index.Name()
	}

	return names
}
