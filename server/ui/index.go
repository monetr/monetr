package ui

import (
	"fmt"
	"html"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

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
