package testutils

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type DebugPrinter struct {
	logger *logrus.Entry
	body   bool
}

// NewDebugPrinter returns a new DebugPrinter given a logger and body
// flag. If body is true, request and response body is also printed.
func NewDebugPrinter(logger *logrus.Entry, body bool) DebugPrinter {
	return DebugPrinter{logger, body}
}

// Request implements Printer.Request.
func (p DebugPrinter) Request(req *http.Request) {
	if req == nil {
		return
	}

	dump, err := httputil.DumpRequest(req, p.body)
	if err != nil {
		panic(err)
	}
	p.logger.Debug("Logging Request\n" + string(dump) + "\n\t")
}

// Response implements Printer.Response.
func (p DebugPrinter) Response(resp *http.Response, duration time.Duration) {
	if resp == nil {
		return
	}

	dump, err := httputil.DumpResponse(resp, p.body)
	if err != nil {
		panic(err)
	}

	text := strings.ReplaceAll(string(dump), "\r\n", "\n")
	lines := strings.SplitN(text, "\n", 2)

	p.logger.Debugf("Logging Response\n%s %s\n%s\t", lines[0], duration, lines[1])
}
