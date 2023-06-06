//go:build monetrbuild

package main

import (
	"os"
	"time"

	"github.com/magefile/mage/mage"
)

func main() {
	code := mage.Main()
	// Give progress time to udpate
	time.Sleep(500 * time.Millisecond)
	os.Exit(code)
}
