package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/monetr/monetr/pkg/client"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	DataCommand = &cobra.Command{
		Use:   "data",
		Short: "Import/export data from your monetr account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	rootCommand.AddCommand(DataCommand)

	newExportDataCommand(DataCommand)
	newImportDataCommand(DataCommand)
}

func newExportDataCommand(parent *cobra.Command) {
	var hostname string
	var token string
	var output string

	command := &cobra.Command{
		Use:   "export",
		Short: "Export data from your monetr account",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logging.NewLoggerWithLevel(config.LogLevel)

			monetrClient := client.NewMonetrHTTPClient(log, hostname, token)

			var err error
			var links []models.Link
			var bankAccounts []models.BankAccount
			var transactions []models.Transaction
			var fundingSchedules []models.FundingSchedule
			var spending []models.Spending

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Minute))
			defer cancel()

			{ // Links
				log.Debug("retrieving links")
				links, err = monetrClient.GetLinks(ctx)
				if err != nil {
					log.WithError(err).Fatalf("failed to retrieve links")
					return err
				}
				log.WithField("count", len(links)).Debug("found links")
			}

			{ // Bank Accounts
				log.Debug("retrieving bank accounts")
				bankAccounts, err = monetrClient.GetBankAccounts(ctx)
				if err != nil {
					log.WithError(err).Fatalf("failed to retrieve bank accounts")
					return err
				}
				log.WithField("count", len(bankAccounts)).Debug("found bank accounts")
			}

			for _, bankAccount := range bankAccounts {
				bankLog := log.WithField("bankAccountId", bankAccount.BankAccountId)

				{ // Funding Schedules
					bankLog.Debug("retrieving funding schedules")
					items, err := monetrClient.GetFundingSchedules(ctx, bankAccount.BankAccountId)
					if err != nil {
						bankLog.WithError(err).Fatalf("failed to retrieve bank accounts")
						return err
					}
					bankLog.WithField("count", len(items)).Debug("found funding schedules")
					fundingSchedules = append(fundingSchedules, items...)
				}

				{ // Spending
					bankLog.Debug("retrieving spending")
					items, err := monetrClient.GetSpending(ctx, bankAccount.BankAccountId)
					if err != nil {
						bankLog.WithError(err).Fatalf("failed to retrieve spending")
						return err
					}
					bankLog.WithField("count", len(items)).Debug("found spending")
					spending = append(spending, items...)
				}

				{ // Transactions
					bankLog.Debug("retrieving transactions")

					var offset int64 = 0
					total := 0
					for {
						items, err := monetrClient.GetTransactions(ctx, bankAccount.BankAccountId, 25, offset)
						if err != nil {
							bankLog.WithError(err).Fatalf("failed to retrieve transactions")
							return err
						}
						transactions = append(transactions, items...)
						offset += 25
						total += len(items)
						if len(items) < 25 {
							break
						}
					}
					bankLog.WithField("count", total).Debug("found transactions")
				}
			}

			me, err := monetrClient.GetMe(ctx)
			if err != nil {
				log.WithError(err).Fatal("failed to retrieve user information")
				return err
			}

			dump := map[string]interface{}{
				"you":              me,
				"links":            links,
				"bankAccounts":     bankAccounts,
				"transactions":     transactions,
				"spending":         spending,
				"fundingSchedules": fundingSchedules,
			}

			dumpRaw, err := json.Marshal(dump)
			if err != nil {
				log.WithError(err).Fatal("failed to encode data export")
				return err
			}

			return errors.Wrap(os.WriteFile(output, dumpRaw, 0644), "failed to write data export")
		},
	}

	command.PersistentFlags().StringVarP(&hostname, "hostname", "H", "https://my.monetr.app", "Specify the hostname (with protocol and port if necessary) of the monetr instance you want to export data from.")
	command.PersistentFlags().StringVarP(&token, "token", "t", "", "Provide your authentication token in order to retrieve the data via HTTP requests to monetr's API.")
	command.PersistentFlags().StringVarP(&output, "output", "o", "monetr_export.json", "Specify an output path for the data export, file will be in a JSON format.")
	_ = command.MarkPersistentFlagRequired("token")
	parent.AddCommand(command)
}

func newImportDataCommand(parent *cobra.Command) {
	var input string
	var dryRun bool

	command := &cobra.Command{
		Use:   "import",
		Short: "Import data from your monetr export into your local monetr instance. This requires database access.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	command.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run the data import, this will print any changes or any failures that would occur during the data import without changing anything.")
	command.PersistentFlags().StringVarP(&input, "input", "i", "monetr_export.json", "Specify the input file, this file must be in the same format as the output of the export command.")
	parent.AddCommand(command)
}
