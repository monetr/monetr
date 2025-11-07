package ui

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type indexParams struct {
	SentryDSN string
}

type indexRenderer struct {
	index *template.Template
}

func (i *indexRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return i.index.Execute(w, data)
}
