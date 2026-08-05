package controller_test

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apiDocsDirectory is where the hand written API reference lives, relative to
// this file. Every route registered in routes.go must show up on one of these
// pages, and every request line on these pages must correspond to a real route.
const apiDocsDirectory = "../../docs/src/v1/en/documentation/api"

// requestLinePattern matches the request line format that every documented
// endpoint uses, for example:
//
//	GET /api/bank_accounts/:bankAccountId/transactions
//
// The format is mandated by docs/API_STYLE.md precisely so it can be parsed
// here. Only lines inside a fenced code block with no language are considered,
// so a method and path mentioned in prose or inside a bash example does not
// accidentally count as documentation.
var requestLinePattern = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) (/\S*)$`)

// TestApiDocumentationCoverage is the entire maintenance story for the API
// reference. Adding a route without documenting it turns the build red, and so
// does deleting a route while leaving its page behind.
//
// If this test fails, the fix is one of:
//   - Document the new endpoint on the appropriate page in docs/src/v1/en/documentation/api,
//     following the template in docs/API_STYLE.md, and check it off in docs/API_COVERAGE.md.
//   - Remove the stale request line from the docs page for a route that no longer exists.
//
// Endpoints that only exist to serve the monetr app still have to appear, but
// they get a deliberately terse entry on the internal endpoints page rather
// than the full template. See docs/API_STYLE.md for what that looks like.
func TestApiDocumentationCoverage(t *testing.T) {
	registered := registeredRoutes(t)
	documented := documentedRoutes(t)

	require.NotEmpty(t, registered, "should have found registered routes, the test harness is broken if this is empty")
	require.NotEmpty(t, documented, "should have found documented routes, the test harness is broken if this is empty")

	undocumented := make([]string, 0, len(registered))
	for route := range registered {
		if _, ok := documented[route]; !ok {
			undocumented = append(undocumented, route)
		}
	}
	sort.Strings(undocumented)

	nonexistent := make([]string, 0, len(documented))
	for route, source := range documented {
		if _, ok := registered[route]; !ok {
			nonexistent = append(nonexistent, fmt.Sprintf("%s (documented in %s)", route, source))
		}
	}
	sort.Strings(nonexistent)

	assert.Emptyf(
		t,
		undocumented,
		"these routes are registered but are not documented anywhere in %s\n\t%s",
		apiDocsDirectory,
		strings.Join(undocumented, "\n\t"),
	)

	assert.Emptyf(
		t,
		nonexistent,
		"these routes are documented but are no longer registered in routes.go\n\t%s",
		strings.Join(nonexistent, "\n\t"),
	)
}

// registeredRoutes builds a controller and registers its routes against a
// throwaway echo instance so the real route table can be read back. Nothing
// here touches the database or any external service, route registration only
// reads configuration.
//
// Feature flags are all turned on deliberately. Some routes are only registered
// when a flag is set, and documentation for them should not disappear just
// because the flag defaults to off.
func registeredRoutes(t *testing.T) map[string]struct{} {
	t.Helper()

	c := &controller.Controller{
		Log: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Configuration: config.Configuration{
			Features: config.Features{
				// Keep every feature flag enabled here. A flag that is off must not
				// let an endpoint slip out of the docs unnoticed.
				TransactionImports: true,
			},
		},
	}

	app := echo.New()
	c.RegisterRoutes(app)

	routes := make(map[string]struct{})
	for _, route := range app.Router().Routes() {
		// echo registers internal routes of its own for each group, using a
		// sentinel method rather than a real HTTP one, to serve 404s for unmatched
		// paths. Those are not endpoints and there is nothing to document about
		// them, so only real HTTP methods count.
		if !documentableMethods[route.Method] {
			continue
		}

		// Only consider what is mounted under the API path.
		if !strings.HasPrefix(route.Path, controller.APIPath) {
			continue
		}

		routes[fmt.Sprintf("%s %s", route.Method, route.Path)] = struct{}{}
	}

	return routes
}

// documentableMethods are the HTTP methods that describe a real endpoint. echo
// uses non-HTTP sentinel values internally (echo.RouteNotFound) which must not
// be mistaken for undocumented routes.
var documentableMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodOptions: true,
}

// documentedRoutes walks the MDX pages and pulls out every request line. The
// returned map points each route at the file it was found in so a stale entry
// can be reported with somewhere to go and fix it.
func documentedRoutes(t *testing.T) map[string]string {
	t.Helper()

	pages, err := filepath.Glob(filepath.Join(apiDocsDirectory, "*.mdx"))
	require.NoError(t, err, "must be able to read the API documentation directory")
	require.NotEmpty(t, pages, "must have found at least one API documentation page")

	routes := make(map[string]string)
	for _, page := range pages {
		for route, line := range requestLinesInFile(t, page) {
			source := fmt.Sprintf("%s:%d", filepath.Base(page), line)
			if existing, ok := routes[route]; ok {
				// Documenting one endpoint twice is its own kind of rot: the two
				// copies drift apart and a reader has no way to tell which is current.
				assert.Failf(
					t,
					"duplicate endpoint documentation",
					"%s is documented in both %s and %s, it should only be documented once",
					route, existing, source,
				)
				continue
			}
			routes[route] = source
		}
	}

	return routes
}

// requestLinesInFile reads a single MDX page and returns the request lines it
// declares. A request line only counts when it is alone inside a fenced code
// block that has no language, which is the format API_STYLE.md requires. That
// rule is what keeps a path mentioned in a curl example or a prose table from
// being mistaken for an endpoint declaration.
func requestLinesInFile(t *testing.T, path string) map[string]int {
	t.Helper()

	file, err := os.Open(path)
	require.NoErrorf(t, err, "must be able to open documentation page %s", path)
	defer file.Close()

	const fenceMarker = "```"

	routes := make(map[string]int)
	var (
		inFence      bool
		fenceHasLang bool
		lineNumber   int
		scanner      = bufio.NewScanner(file)
	)
	// Some pages are long, and the default scanner buffer is not generous.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		lineNumber++
		trimmed := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(trimmed, fenceMarker) {
			if inFence {
				inFence = false
				fenceHasLang = false
				continue
			}

			inFence = true
			// A fence tagged with a language holds an example, not a declaration.
			// ```bash and ```json blocks are full of paths that must not count.
			fenceHasLang = strings.TrimSpace(strings.TrimPrefix(trimmed, fenceMarker)) != ""
			continue
		}

		if !inFence || fenceHasLang {
			continue
		}

		matches := requestLinePattern.FindStringSubmatch(trimmed)
		if matches == nil {
			continue
		}

		routes[fmt.Sprintf("%s %s", matches[1], matches[2])] = lineNumber
	}

	require.NoErrorf(t, scanner.Err(), "must be able to read documentation page %s", path)

	return routes
}
