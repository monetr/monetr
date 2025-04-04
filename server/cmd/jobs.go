package main

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(JobCommand)

	JobCommand.AddCommand(RunJobCommand)
	newCleanupJobsCommand(RunJobCommand)
	RunJobCommand.AddCommand(RunSyncPlaidCommand)
	RunJobCommand.AddCommand(RunRemoveLinkCommand)

	RunSyncPlaidCommand.PersistentFlags().BoolVar(&AllFlag, "all", false, "Pull transactions for all accounts. This job should not be run locally unless you are debugging as it may take a very long time. Will ignore 'account' and 'link' flags.")
	RunSyncPlaidCommand.PersistentFlags().StringVarP(&AccountIDFlag, "account", "a", "", "Account ID to target for the task. Will only run against this account. (required)")
	RunSyncPlaidCommand.PersistentFlags().StringVarP(&LinkIDFlag, "link", "l", "", "Link ID to target for the task. Will only affect objects for this Link. Must belong to the account specified. (required)")
	RunSyncPlaidCommand.PersistentFlags().BoolVarP(&DryRunFlag, "dry-run", "d", false, "Dry run the retrieval of transactions, this will log what transactions might be changed or created. No changes will be persisted. [local]")
	RunSyncPlaidCommand.PersistentFlags().BoolVar(&LocalFlag, "local", false, "Run the job locally, this means the job is not dispatched to the external scheduler like RabbitMQ or Redis. This defaults to true when dry running or when the job engine is in-memory.")

	RunRemoveLinkCommand.PersistentFlags().StringVarP(&LinkIDFlag, "link", "l", "", "Link ID that will be removed forcefully from the database.")
	RunRemoveLinkCommand.PersistentFlags().BoolVarP(&DryRunFlag, "dry-run", "d", false, "Dry run removing the link. No changes will be persisted.")
	RunRemoveLinkCommand.PersistentFlags().BoolVar(&LocalFlag, "local", false, "Run the job locally.")
}

var (
	AllFlag       bool
	AccountIDFlag string
	LinkIDFlag    string
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
		},
	}

	RunSyncPlaidCommand = &cobra.Command{
		Use:   "sync-plaid",
		Short: "Pull latest transactions for a specific link and account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			if AccountIDFlag == "" && !AllFlag {
				return errors.New("--account must be specified if you are not running against --all")
			}
			if LinkIDFlag == "" && AccountIDFlag != "" {
				return errors.New("--link must be specified if you are running against a single account")
			}

			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if AccountIDFlag != "" && AllFlag {
				log.Warn("--account flag does nothing when --all is specified")
			}
			if LinkIDFlag != "" && AllFlag {
				log.Warn("--link flag does nothing when --all is specified")
			}

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				return errors.Wrap(err, "failed to get database instance")
			}

			ctx := context.Background()

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
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				return err
			}

			jobs := make([]background.SyncPlaidArguments, 0)
			if !AllFlag {
				jobArgs := background.SyncPlaidArguments{
					AccountId: models.ID[models.Account](AccountIDFlag),
					LinkId:    models.ID[models.Link](LinkIDFlag),
					Trigger:   "command",
				}
				jobs = append(jobs, jobArgs)
			} else {
				var links []models.Link
				db.Model(&links).
					Where(`"link"."link_type" = ?`, models.PlaidLinkType).
					Where(`"link"."plaid_link_id" IS NOT NULL`).
					Select(&links)
				for _, link := range links {
					jobs = append(jobs, background.SyncPlaidArguments{
						AccountId: link.AccountId,
						LinkId:    link.LinkId,
						Trigger:   "command",
					})
				}
			}

			log.Infof("syncing %d link(s)", len(jobs))

			for _, jobArgs := range jobs {
				if LocalFlag || DryRunFlag {
					txn, err := db.BeginContext(ctx)
					if err != nil {
						log.WithError(err).Fatalf("failed to begin transaction to cleanup jobs")
						return err
					}

					repo := repository.NewRepositoryFromSession(
						clock,
						"user_admin",
						jobArgs.AccountId,
						txn,
						log,
					)

					kms, err := getKMS(log, configuration)
					if err != nil {
						log.WithError(err).Fatal("failed to initialize KMS")
						return err
					}

					secretsRepo := repository.NewSecretsRepository(
						log,
						clock,
						txn,
						kms,
						jobArgs.AccountId,
					)

					job, err := background.NewSyncPlaidJob(
						log,
						repo,
						clock,
						secretsRepo,
						platypus.NewPlaid(log, clock, kms, txn, configuration.Plaid),
						pubsub.NewPostgresPubSub(log, db),
						backgroundJobs,
						jobArgs,
					)
					if err != nil {
						return errors.Wrap(err, "failed to create sync job")
					}

					if err := job.Run(ctx); err != nil {
						log.WithError(err).Fatalf("failed to run sync latest transactions")
						_ = txn.RollbackContext(ctx)
						continue
					}

					if DryRunFlag {
						log.Info("dry run... rolling changes back")
						return txn.RollbackContext(ctx)
					} else {
						return txn.CommitContext(ctx)
					}
				} else {
					return background.TriggerSyncPlaid(ctx, backgroundJobs, jobArgs)
				}
			}

			return nil
		},
	}

	RunRemoveLinkCommand = &cobra.Command{
		Use:   "remove-link",
		Short: "Remove a link from an account, this is a dangerous command and deletes data completely! Be careful!",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			if LinkIDFlag == "" {
				return errors.New("--link must be specified")
			}

			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				return errors.Wrap(err, "failed to get database instance")
			}

			ctx := context.Background()

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
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				return err
			}

			var link models.Link
			if err := db.ModelContext(cmd.Context(), &link).
				Where(`"link"."link_id" = ?`, LinkIDFlag).
				Limit(1).
				Select(&link); err != nil {
				return errors.Wrap(err, "failed to retrieve link specified")
			}

			jobArgs := background.RemoveLinkArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
			}

			if DryRunFlag {
				log.Info("dry run removing link")
			}

			if LocalFlag || DryRunFlag {
				txn, err := db.BeginContext(ctx)
				if err != nil {
					log.WithError(err).Fatalf("failed to begin transaction to remove link")
					return err
				}

				job, err := background.NewRemoveLinkJob(
					log,
					txn,
					clock,
					pubsub.NewPostgresPubSub(log, db),
					jobArgs,
				)
				if err != nil {
					return errors.Wrap(err, "failed to create remove link job")
				}

				if err := job.Run(ctx); err != nil {
					log.WithError(err).Fatalf("failed to run remove link job")
					_ = txn.RollbackContext(ctx)
					return err
				}

				if DryRunFlag {
					log.Info("dry run... rolling changes back")
					return txn.RollbackContext(ctx)
				} else {
					return txn.CommitContext(ctx)
				}
			} else {
				return background.TriggerRemoveLink(ctx, backgroundJobs, jobArgs)
			}
		},
	}
)
