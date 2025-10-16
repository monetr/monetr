//go:build unix

package zoneinfo

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	zoneInfoLocations = []string{
		// Not sure where they are on other OSes. This is accurate for Debian though
		// which is what monetr runs in inside a container anyway.
		"/usr/share/zoneinfo/tzdata.zi",
	}
	loadAliasesOnce = sync.Once{}
)

func LoadAliasesFromHost(ctx context.Context, logEntry *logrus.Entry) {
	loadAliasesOnce.Do(func() {
		log := logEntry.WithContext(ctx)
		for _, zoneInfoLocation := range zoneInfoLocations {
			if err := ParseAliasesFromFile(zoneInfoLocation, aliases); err != nil {
				log.WithField("filename", zoneInfoLocation).
					WithError(err).
					Warn("failed to parse zone info from file")
			}
		}
	})
}

// ParseAliasesFromFile parses a file in the zic file format but is looking
// specifically for Link data inside the file. All other information is ignored.
// See https://man7.org/linux/man-pages/man8/zic.8.html for more information.
func ParseAliasesFromFile(path string, aliasMap map[string]string) error {
	reader, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer reader.Close()

	buffer := bufio.NewReader(reader)
	var currentLine []byte
	for {
		line, isPrefix, err := buffer.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrap(err, "failed to parse aliases from tzdata file")
		}

		currentLine = append(currentLine, line...)
		if isPrefix {
			continue
		}

		// If the current tzdata line is not a link then we should clear our current
		// line and move on. We are looking for only the link lines
		if currentLine[0] != 'L' {
			currentLine = nil
			continue
		}

		parts := bytes.SplitAfterN(currentLine, []byte(" "), 3)

		var oldName, newName []byte
		for i, part := range parts {
			switch i {
			case 0: // Link part of the link, no-op
			case 1: // New name
				newName = bytes.TrimSpace(part)
			case 2: // Old name
				oldName = bytes.TrimSpace(part)
			}
		}

		aliasMap[string(oldName)] = string(newName)

		currentLine = nil
	}

	return nil
}
