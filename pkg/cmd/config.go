package main

import (
	"path/filepath"

	"github.com/monetr/monetr/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	configOutputPath string
)

func init() {
	rootCommand.AddCommand(ConfigCommand)
	ConfigCommand.AddCommand(GenerateConfigCommand)

	GenerateConfigCommand.PersistentFlags().StringVarP(&configOutputPath, "output", "o", "./config.yaml", "Specify the output path of the generated config, defaults to ./config.yaml")
}

var (
	ConfigCommand = &cobra.Command{
		Use:   "config",
		Short: "Some configuration helper commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	GenerateConfigCommand = &cobra.Command{
		Use:   "generate",
		Short: "Generate a config.yaml using default values.",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputConfig, err := filepath.Abs(configOutputPath)
			if err != nil {
				return errors.Wrap(err, "failed to determine the absolute path of the generated config file")
			}

			if err = config.GenerateConfigFile(&config.FilePath, outputConfig); err != nil {
				return errors.Wrap(err, "failed to generate the config file")
			}

			return nil
		},
	}
)
