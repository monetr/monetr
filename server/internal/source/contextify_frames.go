/*

FROM: https://github.com/getsentry/sentry-go/blob/df20ce63bbede6de539d2ad2f15f7daf2b703c65/integrations.go#L215-L328

MIT License

Copyright (c) 2019 Functional Software, Inc. dba Sentry

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package source

import (
	"sync"

	"github.com/getsentry/sentry-go"
)

// ================================
// Contextify Frames Integration
// ================================

type ContextifyFramesIntegration struct {
	sr              sourceReader
	contextLines    int
	cachedLocations sync.Map
}

func (cfi *ContextifyFramesIntegration) Name() string {
	return "ContextifyFrames"
}

func (cfi *ContextifyFramesIntegration) SetupOnce(client *sentry.Client) {
	cfi.sr = newSourceReader()
	cfi.contextLines = 5

	client.AddEventProcessor(cfi.processor)
}

func (cfi *ContextifyFramesIntegration) processor(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	// Range over all exceptions
	for _, ex := range event.Exception {
		// If it has no stacktrace, just bail out
		if ex.Stacktrace == nil {
			continue
		}

		// If it does, it should have frames, so try to contextify them
		ex.Stacktrace.Frames = cfi.contextify(ex.Stacktrace.Frames)
	}

	// Range over all threads
	for _, th := range event.Threads {
		// If it has no stacktrace, just bail out
		if th.Stacktrace == nil {
			continue
		}

		// If it does, it should have frames, so try to contextify them
		th.Stacktrace.Frames = cfi.contextify(th.Stacktrace.Frames)
	}

	return event
}

func (cfi *ContextifyFramesIntegration) contextify(frames []sentry.Frame) []sentry.Frame {
	contextifiedFrames := make([]sentry.Frame, 0, len(frames))

	for _, frame := range frames {
		if !frame.InApp {
			contextifiedFrames = append(contextifiedFrames, frame)
			continue
		}

		path := frame.AbsPath
		if path == "" {
			contextifiedFrames = append(contextifiedFrames, frame)
			continue
		}

		lines, contextLine := cfi.sr.readContextLines(path, frame.Lineno, cfi.contextLines)
		contextifiedFrames = append(contextifiedFrames, cfi.addContextLinesToFrame(frame, lines, contextLine))
	}

	return contextifiedFrames
}

func (cfi *ContextifyFramesIntegration) addContextLinesToFrame(frame sentry.Frame, lines [][]byte, contextLine int) sentry.Frame {
	for i, line := range lines {
		switch {
		case i < contextLine:
			frame.PreContext = append(frame.PreContext, string(line))
		case i == contextLine:
			frame.ContextLine = string(line)
		default:
			frame.PostContext = append(frame.PostContext, string(line))
		}
	}
	return frame
}
