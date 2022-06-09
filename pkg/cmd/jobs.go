package main

import (
	"context"
	"fmt"
	"time"

	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/vault_helper"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(JobCommand)

	JobCommand.AddCommand(RunJobCommand)
	newCleanupJobsCommand(RunJobCommand)
	RunJobCommand.AddCommand(RunPullTransactionsCommand)

	RunPullTransactionsCommand.PersistentFlags().BoolVar(&AllFlag, "all", false, "Pull transactions for all accounts. This job should not be run locally unless you are debugging as it may take a very long time. Will ignore 'account' and 'link' flags.")
	RunPullTransactionsCommand.PersistentFlags().Uint64VarP(&AccountIDFlag, "account", "a", 0, "Account ID to target for the task. Will only run against this account. (required)")
	RunPullTransactionsCommand.PersistentFlags().Uint64VarP(&LinkIDFlag, "link", "l", 0, "Link ID to target for the task. Will only affect objects for this Link. Must belong to the account specified. (required)")
	RunPullTransactionsCommand.PersistentFlags().DurationVarP(&SinceFlag, "since", "s", 0, "Retrieve transactions since duration. For example, '7d' will retrieve the past 7 days.")
	RunPullTransactionsCommand.PersistentFlags().BoolVarP(&DryRunFlag, "dry-run", "d", false, "Dry run the retrieval of transactions, this will log what transactions might be changed or created. No changes will be persisted. [local]")
	RunPullTransactionsCommand.PersistentFlags().BoolVar(&LocalFlag, "local", false, "Run the job locally, this means the job is not dispatched to the external scheduler like RabbitMQ or Redis. This defaults to true when dry running or when the job engine is in-memory.")
}

var (
	AllFlag       bool
	AccountIDFlag uint64
	LinkIDFlag    uint64
	SinceFlag     time.Duration
	DryRunFlag    bool
	LocalFlag     bool
)

var (
	JobCommand = &cobra.Command{
		Use:   "jobs",
		Short: "Trigger jobs to be run by monetr instances or by this instance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	RunJobCommand = &cobra.Command{
		Use:   "run [flags] [command]",
		Short: "Run a specific job.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()

			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := getDatabase(log, configuration, nil)

			redisController, err := cache.NewRedisCache(log, configuration.Redis)
			if err != nil {
				log.WithError(err).Fatalf("failed to create redis cache: %+v", err)
				return err
			}
			defer redisController.Close()

			backgroundJobs, err := background.NewBackgroundJobs(
				cmd.Context(),
				log,
				configuration,
				db,
				redisController.Pool(),
				nil,
				nil,
				nil,
			)
			if err != nil {
				return err
			}

			triggerableJobNames := backgroundJobs.GetTriggerableJobNames()
			commands := map[string]*cobra.Command{}
			for _, jobName := range triggerableJobNames {
				name := jobName
				command := &cobra.Command{
					Use:   jobName,
					Short: "Trigger the job",
					RunE: func(cmdInner *cobra.Command, argsInner []string) error {
						fmt.Println(name)

						return nil
					},
				}

				commands[jobName] = command
				cmd.AddCommand(command)
			}

			switch len(args) {
			case 0:
				return cmd.Help()
			case 1:
				command, ok := commands[args[0]]
				if !ok {
					return cmd.Help()
				}

				return command.Execute()
			default:
				return cmd.Help()
			}
		},
	}

	RunPullTransactionsCommand = &cobra.Command{
		Use:   "pull-latest-transactions",
		Short: "Pull latest transactions for a specific link and account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if AccountIDFlag == 0 && !AllFlag {
				return errors.New("--account must be specified if you are not running against --all")
			}
			if LinkIDFlag == 0 && AccountIDFlag != 0 {
				return errors.New("--link must be specified if you are running against a single account")
			}

			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if AccountIDFlag != 0 && AllFlag {
				log.Warn("--account flag does nothing when --all is specified")
			}
			if LinkIDFlag != 0 && AllFlag {
				log.Warn("--link flag does nothing when --all is specified")
			}

			db, err := getDatabase(log, configuration, nil)

			ctx := context.Background()

			if SinceFlag == 0 {
				SinceFlag = 24 * time.Hour
			}

			jobArgs := background.PullTransactionsArguments{
				AccountId: AccountIDFlag,
				LinkId:    LinkIDFlag,
				Start:     time.Now().Add(-SinceFlag),
				End:       time.Now(),
			}

			if LocalFlag || DryRunFlag {
				log.Info("running locally")

				var vault vault_helper.VaultHelper
				if configuration.Vault.Enabled {
					log.Debug("vault is enabled for secret storage")
					client, err := vault_helper.NewVaultHelper(log, vault_helper.Config{
						Address:         configuration.Vault.Address,
						Role:            configuration.Vault.Role,
						Auth:            configuration.Vault.Auth,
						Token:           configuration.Vault.Token,
						TokenFile:       configuration.Vault.TokenFile,
						Timeout:         configuration.Vault.Timeout,
						IdleConnTimeout: configuration.Vault.IdleConnTimeout,
						Username:        configuration.Vault.Username,
						Password:        configuration.Vault.Password,
					})
					if err != nil {
						log.WithError(err).Fatalf("failed to create vault helper")
						return err
					}

					vault = client
				}

				txn, err := db.BeginContext(ctx)
				if err != nil {
					log.WithError(err).Fatalf("failed to begin transaction to cleanup jobs")
					return err
				}

				repo := repository.NewRepositoryFromSession(0, AccountIDFlag, txn)
				var plaidSecrets secrets.PlaidSecretsProvider
				if configuration.Vault.Enabled {
					log.Debugf("secrets will be stored in vault")
					plaidSecrets = secrets.NewVaultPlaidSecretsProvider(log, vault)
				} else {
					log.Debugf("secrets will be stored in postgres")
					plaidSecrets = secrets.NewPostgresPlaidSecretsProvider(log, db)
				}

				job, err := background.NewPullTransactionsJob(
					log,
					repo,
					plaidSecrets,
					platypus.NewPlaid(log, plaidSecrets, repository.NewPlaidRepository(txn), configuration.Plaid),
					pubsub.NewPostgresPubSub(log, db),
					jobArgs,
				)

				if err := job.Run(ctx); err != nil {
					log.WithError(err).Fatalf("failed to run pull latest transactions")
					_ = txn.RollbackContext(ctx)
					return err
				}

				if DryRunFlag {
					log.Info("dry run... rolling changes back")
					return txn.RollbackContext(ctx)
				} else {
					return txn.CommitContext(ctx)
				}
			}

			redisController, err := cache.NewRedisCache(log, configuration.Redis)
			if err != nil {
				log.WithError(err).Fatalf("failed to create redis cache: %+v", err)
				return err
			}
			defer redisController.Close()

			backgroundJobs, err := background.NewBackgroundJobs(
				cmd.Context(),
				log,
				configuration,
				db,
				redisController.Pool(),
				nil,
				nil,
				nil,
			)
			if err != nil {
				return err
			}

			return background.TriggerPullTransactions(ctx, backgroundJobs, jobArgs)
		},
	}
)
