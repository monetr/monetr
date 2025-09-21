package commands

import (
	"github.com/monetr/monetr/server/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RegisterGlobalFlags(parent *cobra.Command) {
	parent.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "L", "info", "Specify the log level to use, allowed values: trace, debug, info, warn, error, fatal")
	parent.PersistentFlags().StringArrayVarP(&config.FilePath, "config", "c", []string{}, "Specify the config file to use.")
	viper.BindPFlag("Logging.Level", parent.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("configFile", parent.PersistentFlags().Lookup("config"))
}
