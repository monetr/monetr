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

	GenerateConfigCommand.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Specify a config file to use as a template, if omitted ./config.yaml, ~/.monetr/config.yaml or /etc/monetr/config.yaml will be used. This file will not be overwritten.")
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
			var configPath *string
			if len(configFilePath) > 0 {
				configPath = &configFilePath

				inputConfig, err := filepath.Abs(configFilePath)
				if err != nil {
					return errors.Wrap(err, "failed to determine the absolute path of the input config file")
				}

				outputConfig, err := filepath.Abs(configOutputPath)
				if err != nil {
					return errors.Wrap(err, "failed to determine the absolute path of the generated config file")
				}

				if inputConfig == outputConfig {
					return errors.New("input config and output cannot be the same, monetr will not overwrite the input file")
				}
			}

			if err := config.GenerateConfigFile(configPath, configOutputPath); err != nil {
				return errors.Wrap(err, "failed to generate the config file")
			}

			return nil
		},
	}
)
