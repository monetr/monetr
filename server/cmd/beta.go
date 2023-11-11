package main

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/hash"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCommand.AddCommand(BetaCommand)
	BetaCommand.AddCommand(NewBetaCodeCommand)

	BetaCommand.PersistentFlags().StringVarP(&postgresAddress, "host", "H", "localhost", "PostgreSQL host address.")
	BetaCommand.PersistentFlags().IntVarP(&postgresPort, "port", "P", 5432, "PostgreSQL port.")
	BetaCommand.PersistentFlags().StringVarP(&postgresUsername, "username", "U", "postgres", "PostgreSQL user.")
	BetaCommand.PersistentFlags().StringVarP(&postgresPassword, "password", "W", "", "PostgreSQL password.")
	BetaCommand.PersistentFlags().StringVarP(&postgresDatabase, "database", "d", "postgres", "PostgreSQL database.")
	// TODO This doesn't account for TLS properties that would need to be set.
	viper.BindPFlag("PostgreSQL.Address", BetaCommand.PersistentFlags().Lookup("host"))
	viper.BindPFlag("PostgreSQL.Port", BetaCommand.PersistentFlags().Lookup("port"))
	viper.BindPFlag("PostgreSQL.Username", BetaCommand.PersistentFlags().Lookup("username"))
	viper.BindPFlag("PostgreSQL.Password", BetaCommand.PersistentFlags().Lookup("password"))
	viper.BindPFlag("PostgreSQL.Database", BetaCommand.PersistentFlags().Lookup("database"))
}

var (
	BetaCommand = &cobra.Command{
		Use:   "beta",
		Short: "Manage beta things",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}
			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to establish database connection")
				return err
			}
			defer db.Close()

			betas := make([]models.Beta, 0)
			if err := db.Model(&betas).Select(&betas); err != nil {
				log.WithError(err).Error("failed to retrieve beta(s)")
				return errors.Wrap(err, "failed to retrieve beta(s)")
			}

			log.Infof("found %d beta(s)", len(betas))

			var used, unused, expired uint32
			now := time.Now()
			for _, beta := range betas {
				if beta.UsedByUserId != nil {
					used++
				} else if now.Before(beta.ExpiresAt) {
					unused++
				} else {
					expired++
				}
			}
			log.Infof("used: %d", used)
			log.Infof("unused: %d", unused)
			log.Infof("expired: %d", expired)

			return nil
		},
	}

	NewBetaCodeCommand = &cobra.Command{
		Use:   "new-code",
		Short: "Generates a beta code and returns the code, the code is encrypted then added to the database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}
			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to establish database connection")
				return err
			}
			defer db.Close()

			random := make([]byte, 8)
			_, err = rand.Read(random)
			if err != nil {
				log.WithError(err).Error("failed to read random data")
				return errors.Wrap(err, "failed to read random data")
			}

			betaCode := fmt.Sprintf("%X-%X", random[:4], random[4:])

			expires := util.Midnight(time.Now().Add(14*24*time.Hour), time.Local)
			beta := models.Beta{
				CodeHash:  hash.HashPassword(strings.ToLower(betaCode), betaCode),
				ExpiresAt: expires,
			}

			if _, err := db.Model(&beta).Insert(&beta); err != nil {
				log.WithError(err).Error("failed to generate beta code")
				return errors.Wrap(err, "failed to generate beta code")
			}

			fmt.Println("NEW BETA CODE:")
			fmt.Println()
			fmt.Println(betaCode)
			fmt.Println()
			fmt.Println("Code expires on: ", expires)

			return nil
		},
	}
)
