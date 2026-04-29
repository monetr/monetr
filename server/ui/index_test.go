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

func TestIsImmutableAssetPath(t *testing.T) {
	t.Run("script asset is immutable", func(t *testing.T) {
		assert.True(t, isImmutableAssetPath("/assets/scripts/abc12345.js"))
	})

	t.Run("style asset is immutable", func(t *testing.T) {
		assert.True(t, isImmutableAssetPath("/assets/styles/abc12345.css"))
	})

	t.Run("font asset is immutable", func(t *testing.T) {
		assert.True(t, isImmutableAssetPath("/assets/fonts/inter.abc12345.woff2"))
	})

	t.Run("nested script asset is immutable", func(t *testing.T) {
		assert.True(t, isImmutableAssetPath("/assets/scripts/chunks/abc12345.js"))
	})

	t.Run("resources asset is not immutable", func(t *testing.T) {
		// The resources directory is not content hashed by rsbuild, so its contents
		// must remain revalidatable.
		assert.False(t, isImmutableAssetPath("/assets/resources/logo.png"))
	})

	t.Run("images asset is not immutable", func(t *testing.T) {
		assert.False(t, isImmutableAssetPath("/assets/images/logo.png"))
	})

	t.Run("manifest at root is not immutable", func(t *testing.T) {
		assert.False(t, isImmutableAssetPath("/manifest.json"))
	})

	t.Run("robots at root is not immutable", func(t *testing.T) {
		assert.False(t, isImmutableAssetPath("/robots.txt"))
	})

	t.Run("bare assets prefix is not immutable", func(t *testing.T) {
		assert.False(t, isImmutableAssetPath("/assets/"))
	})

	t.Run("unrelated path is not immutable", func(t *testing.T) {
		assert.False(t, isImmutableAssetPath("/some/other/path.js"))
	})

	t.Run("empty path is not immutable", func(t *testing.T) {
		assert.False(t, isImmutableAssetPath(""))
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
