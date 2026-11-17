package commands

import "github.com/spf13/cobra"

func EncryptionCommand(parent *cobra.Command) {
	encryptionCommand := &cobra.Command{
		Use:   "encryption",
		Short: "Manage and maintain monetr's built in at rest encryption",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	parent.AddCommand(encryptionCommand)
}
