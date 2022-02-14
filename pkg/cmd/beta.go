package main

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(BetaCommand)
	BetaCommand.AddCommand(NewBetaCodeCommand)

	BetaCommand.PersistentFlags().StringVarP(&postgresAddress, "host", "H", "localhost", "PostgreSQL host address.")
	BetaCommand.PersistentFlags().IntVarP(&postgresPort, "port", "P", 5432, "PostgreSQL port.")
	BetaCommand.PersistentFlags().StringVarP(&postgresUsername, "username", "U", "postgres", "PostgreSQL user.")
	BetaCommand.PersistentFlags().StringVarP(&postgresPassword, "password", "W", "", "PostgreSQL password.")
	BetaCommand.PersistentFlags().StringVarP(&postgresDatabase, "database", "d", "postgres", "PostgreSQL database.")
}

var (
	BetaCommand = &cobra.Command{
		Use:   "beta",
		Short: "Manage beta things",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logging.NewLogger()

			options := getDatabaseCommandConfiguration()

			db := pg.Connect(options)
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
			log := logging.NewLogger()

			options := getDatabaseCommandConfiguration()

			db := pg.Connect(options)
			defer db.Close()

			random := make([]byte, 8)
			_, err := rand.Read(random)
			if err != nil {
				log.WithError(err).Error("failed to read random data")
				return errors.Wrap(err, "failed to read random data")
			}

			betaCode := fmt.Sprintf("%X-%X", random[:4], random[4:])

			expires := util.MidnightInLocal(time.Now().Add(14*24*time.Hour), time.Local)
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
