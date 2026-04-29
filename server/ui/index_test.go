package ui

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPreconnectTag(t *testing.T) {
	t.Run("empty origin returns empty html", func(t *testing.T) {
		assert.Equal(t, template.HTML(""), buildPreconnectTag(""))
	})

	t.Run("https origin produces a preconnect tag with crossorigin", func(t *testing.T) {
		got := buildPreconnectTag("https://o12345.ingest.sentry.io")
		assert.Equal(
			t,
			template.HTML(`<link rel="preconnect" href="https://o12345.ingest.sentry.io" crossorigin />`),
			got,
		)
	})

	t.Run("origin with port is preserved verbatim", func(t *testing.T) {
		got := buildPreconnectTag("https://sentry.internal.example.com:8443")
		assert.Equal(
			t,
			template.HTML(`<link rel="preconnect" href="https://sentry.internal.example.com:8443" crossorigin />`),
			got,
		)
	})

	t.Run("origin is html escaped defensively", func(t *testing.T) {
		// Origins coming back from Sentry.GetExternalOrigin are constrained to
		// scheme+host, but the defensive escape pass should still neutralize a
		// stray quote rather than emitting attribute-breaking html.
		got := buildPreconnectTag(`https://o12345.ingest.sentry.io"&<>`)
		assert.Equal(
			t,
			template.HTML(`<link rel="preconnect" href="https://o12345.ingest.sentry.io&#34;&amp;&lt;&gt;" crossorigin />`),
			got,
		)
	})
}

// TestIndexRenderer_PreconnectTag exercises the renderer with an inline
// template instead of the embedded UI filesystem so that the wiring between
// indexParams and the rendered output is verifiable without a built UI.
func TestIndexRenderer_PreconnectTag(t *testing.T) {
	tmpl := template.Must(template.New("index").Parse(
		`<head>{{ .PreconnectTag }}</head>`,
	))
	renderer := &indexRenderer{
		index: tmpl,
	}

	cases := []struct {
		name     string
		params   indexParams
		expected string
	}{
		{
			name:     "empty preconnect tag is omitted from head",
			params:   indexParams{},
			expected: `<head></head>`,
		},
		{
			name: "populated preconnect tag is rendered as raw html",
			params: indexParams{
				PreconnectTag: buildPreconnectTag("https://o12345.ingest.sentry.io"),
			},
			expected: `<head><link rel="preconnect" href="https://o12345.ingest.sentry.io" crossorigin /></head>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			require.NoError(t, renderer.Render(&buf, "", tc.params, nil))
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}
