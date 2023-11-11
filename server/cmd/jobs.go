package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
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
			clock := clock.New()

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
				clock,
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
			clock := clock.New()
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
				Start:     clock.Now().Add(-SinceFlag),
				End:       clock.Now(),
			}

			if LocalFlag || DryRunFlag {
				log.Info("running locally")
				txn, err := db.BeginContext(ctx)
				if err != nil {
					log.WithError(err).Fatalf("failed to begin transaction to cleanup jobs")
					return err
				}

				repo := repository.NewRepositoryFromSession(clock, 0, AccountIDFlag, txn)

				kms, err := getKMS(log, configuration)
				if err != nil {
					log.WithError(err).Fatal("failed to initialize KMS")
					return err
				}

				plaidSecrets := secrets.NewPostgresPlaidSecretsProvider(log, db, kms)
				job, err := background.NewPullTransactionsJob(
					log,
					repo,
					clock,
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
				clock,
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
