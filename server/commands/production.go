//go:build !development

package commands

import "github.com/spf13/cobra"

func DevelopmentCommand(*cobra.Command) {
	// No-op
}
