package commands

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func jobRemoveLink(parent *cobra.Command) {
	var arguments struct {
		LinkID string
		DryRun bool
		Local  bool
	}

	command := &cobra.Command{
		Use:   "remove-link",
		Short: "Remove a link from an account, this is a dangerous command and deletes data completely! Be careful!",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			if arguments.LinkID == "" {
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
				Where(`"link"."link_id" = ?`, arguments.LinkID).
				Limit(1).
				Select(&link); err != nil {
				return errors.Wrap(err, "failed to retrieve link specified")
			}

			jobArgs := background.RemoveLinkArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
			}

			if arguments.DryRun {
				log.Info("dry run removing link")
			}

			if arguments.Local || arguments.DryRun {
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

				if arguments.DryRun {
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

	command.PersistentFlags().StringVarP(&arguments.LinkID, "link", "l", "", "Link ID that will be removed forcefully from the database.")
	command.PersistentFlags().BoolVarP(&arguments.DryRun, "dry-run", "d", false, "Dry run removing the link. No changes will be persisted.")
	command.PersistentFlags().BoolVar(&arguments.Local, "local", false, "Run the job locally.")

	parent.AddCommand(command)
}
