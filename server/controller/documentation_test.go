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
// this file.
const apiDocsDirectory = "../../docs/src/v1/en/documentation/api"

// requestLinePattern matches how a documented endpoint declares itself: `GET
// /api/bank_accounts/:bankAccountId/transactions` inside a ```http fence. The
// API documentation guide in documentation.mdx requires that format so it can
// be parsed here.
var requestLinePattern = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) (/\S*)$`)

// TestApiDocumentationCoverage turns the build red when a route is added
// without docs, or deleted while its page stays behind. To fix a failure:
// document the endpoint following the API documentation guide in
// documentation.mdx, or delete the stale request line.
//
// Internal endpoints still have to appear, they just get a terse entry.
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

// registeredRoutes registers the real routes against a throwaway echo so the
// table can be read back. Touches no database, registration only reads config.
func registeredRoutes(t *testing.T) map[string]struct{} {
	t.Helper()

	c := &controller.Controller{
		Log: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Configuration: config.Configuration{
			Features: config.Features{
				// All flags on. A flag that is off must not let an endpoint slip out.
				TransactionImports: true,
			},
		},
	}

	app := echo.New()
	c.RegisterRoutes(app)

	routes := make(map[string]struct{})
	for _, route := range app.Router().Routes() {
		// echo adds its own 404 routes with a sentinel method. Not endpoints.
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

// documentableMethods are the real HTTP methods. echo's internal sentinels are
// not endpoints.
var documentableMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodOptions: true,
}

// documentedRoutes returns every request line in the MDX pages, mapped to where
// it was found so a stale one can be reported with a file and line.
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
				// Two copies drift apart and nobody can tell which is current.
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

// requestLinesInFile returns the request lines one page declares. Only counts a
// line alone inside a ```http fence, so paths in curl examples and prose tables
// don't get mistaken for one.
func requestLinesInFile(t *testing.T, path string) map[string]int {
	t.Helper()

	file, err := os.Open(path)
	require.NoErrorf(t, err, "must be able to open documentation page %s", path)
	defer file.Close()

	const fenceMarker = "```"

	routes := make(map[string]int)
	var (
		inFence        bool
		fenceIsRequest bool
		lineNumber     int
		scanner        = bufio.NewScanner(file)
	)
	// Pages are long and the default buffer is small.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		lineNumber++
		trimmed := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(trimmed, fenceMarker) {
			if inFence {
				inFence = false
				fenceIsRequest = false
				continue
			}

			inFence = true
			// ```http for highlighting. Bare fences still count. Anything else is an
			// example: ```bash and ```json are full of paths that must not count.
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, fenceMarker))
			fenceIsRequest = lang == "" || lang == "http"
			continue
		}

		if !inFence || !fenceIsRequest {
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
