package commands

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func adminRegisterCode(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "register:code",
		Short: "Generate a new registration code.",
		Long:  "This is used to generate beta codes when they are enabled to restrict access to sign ups for the monetr application. This will generate a temporary code and return it to be used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}
			db, err := database.GetDatabase(log, configuration, nil)
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

			hash := sha256.New()
			hash.Write([]byte(strings.ToLower(betaCode)))
			hashedCode := fmt.Sprintf("%X", hash.Sum(nil))

			beta := models.Beta{
				CodeHash:  hashedCode,
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

	parent.AddCommand(command)
}
