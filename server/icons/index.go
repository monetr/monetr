package icons

import "sync"

type IconIndex interface {
	Search(input string) *Icon
	Name() string
}

var (
	enabled bool
	// loaders are registered by each icon pack's init function. They are not
	// invoked until something actually needs icons, building an index can be
	// expensive and most commands (like status) never use them.
	loaders  []func() IconIndex
	loadOnce sync.Once
	indexes  []IconIndex
)

func registerIndex(loader func() IconIndex) {
	loaders = append(loaders, loader)
}

func load() {
	loadOnce.Do(func() {
		for _, loader := range loaders {
			if index := loader(); index != nil {
				indexes = append(indexes, index)
			}
		}
	})
}

// Load builds the icon indexes if they have not been built yet. It is safe to
// call concurrently, callers that race the first build will block until it is
// finished.
func Load() {
	load()
}

func GetIconsEnabled() bool {
	if !enabled {
		return false
	}

	load()
	return len(indexes) > 0
}

func GetIconIndexes() []string {
	load()
	names := make([]string, len(indexes))
	for i, index := range indexes {
		names[i] = index.Name()
	}

	return names
}
