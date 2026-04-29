package ui

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
)

// immutableAssetPrefixes lists the rsbuild output directories whose filenames
// embed a content hash, which means the bytes for a given URL never change.
// /assets/resources is intentionally excluded because resource files are not
// content hashed and would serve stale bytes if marked immutable.
var immutableAssetPrefixes = []string{
	"/assets/scripts/",
	"/assets/styles/",
	"/assets/fonts/",
}

type indexParams struct {
	SentryDSN     string
	PreconnectTag template.HTML
}

type indexRenderer struct {
	index *template.Template
}

func (i *indexRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return i.index.Execute(w, data)
}

// buildPreconnectTag will return a fully formed preconnect link tag for the
// given origin, or an empty html fragment when the origin is empty. The origin
// is html escaped defensively even though the upstream helper that produces it
// already constrains the value to scheme and host.
func buildPreconnectTag(origin string) template.HTML {
	if origin == "" {
		return ""
	}
	return template.HTML(fmt.Sprintf(
		`<link rel="preconnect" href="%s" crossorigin />`,
		html.EscapeString(origin),
	))
}

// isImmutableAssetPath will return true when the given request path lives in
// one of the rsbuild output directories that uses content hashing in the
// filename, and is therefore safe to mark immutable in the Cache-Control
// header.
func isImmutableAssetPath(p string) bool {
	for _, prefix := range immutableAssetPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}
