package background

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

const (
	ProcessSpendingName = "ProcessSpending"
)

type (
	ProcessSpendingHandler struct {
		log          *slog.Logger
		db           *pg.DB
		repo         repository.JobRepository
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	ProcessSpendingArguments struct {
		AccountId     ID[Account]     `json:"accountId"`
		BankAccountId ID[BankAccount] `json:"bankAccountId"`
	}

	ProcessSpendingJob struct {
		args  ProcessSpendingArguments
		log   *slog.Logger
		repo  repository.BaseRepository
		clock clock.Clock
	}
)

func NewProcessSpendingHandler(
	log *slog.Logger,
	db *pg.DB,
	clock clock.Clock,
) *ProcessSpendingHandler {
	return &ProcessSpendingHandler{
		log:          log,
		db:           db,
		repo:         repository.NewJobRepository(db, clock),
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (p ProcessSpendingHandler) QueueName() string {
	return ProcessSpendingName
}

func (p *ProcessSpendingHandler) HandleConsumeJob(
	ctx context.Context,
	log *slog.Logger,
	data []byte,
) error {
	var args ProcessSpendingArguments
	if err := errors.Wrap(p.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Process Spending job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log = log.With(
			"accountId", args.AccountId,
			"bankAccountId", args.BankAccountId,
		)

		repo := repository.NewRepositoryFromSession(
			p.clock,
			"user_system",
			args.AccountId,
			txn,
			log,
		)
		job, err := NewProcessSpendingJob(
			log,
			repo,
			args,
			p.clock,
		)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func (p ProcessSpendingHandler) DefaultSchedule() string {
	// Run once an hour at minute 30.
	return "0 30 * * * *"
}

func (p *ProcessSpendingHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := p.log

	log.InfoContext(ctx, "retrieving bank accounts with stale spending")
	bankAccountsWithStaleSpending, err := p.repo.GetBankAccountsWithStaleSpending(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve bank accounts with stale spending")
	}

	if len(bankAccountsWithStaleSpending) == 0 {
		crumbs.Debug(ctx, "No bank accounts had stale spending objects.", nil)
		log.InfoContext(ctx, "no bank accounts have stale spending objects")
		return nil
	}

	log.InfoContext(ctx, "found bank accounts with stale spending", "count", len(bankAccountsWithStaleSpending))
	crumbs.Debug(ctx, "Found bank accounts with stale spending.", map[string]any{
		"count": len(bankAccountsWithStaleSpending),
	})

	jobErrors := make([]error, 0)

	for _, item := range bankAccountsWithStaleSpending {
		itemLog := log.With(
			"accountId", item.AccountId,
			"bankAccountId", item.BankAccountId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing bank account to process stale spending")
		err = enqueuer.EnqueueJob(ctx, p.QueueName(), ProcessSpendingArguments{
			AccountId:     item.AccountId,
			BankAccountId: item.BankAccountId,
		})
		if err != nil {
			log.WarnContext(ctx, "failed to enqueue job to process stale spending", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to process stale spending", "job", map[string]any{
				"error": err,
			})
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued bank accounts for stale spending processing")
	}

	return nil
}

func NewProcessSpendingJob(
	log *slog.Logger,
	repo repository.BaseRepository,
	args ProcessSpendingArguments,
	clock clock.Clock,
) (*ProcessSpendingJob, error) {
	return &ProcessSpendingJob{
		args:  args,
		log:   log,
		repo:  repo,
		clock: clock,
	}, nil
}

func (p *ProcessSpendingJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := p.log

	account, err := p.repo.GetAccount(span.Context())
	if err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve account to process stale spending", "err", err)
		return err
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		log.ErrorContext(span.Context(), "failed to parse account timezone", "err", err)
		return err
	}

	now := p.clock.Now()
	allSpending, err := p.repo.GetSpending(span.Context(), p.args.BankAccountId)
	if err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve spending for bank account", "err", err)
		return err
	}

	fundingSchedules := map[ID[FundingSchedule]]*FundingSchedule{}

	spendingToUpdate := make([]Spending, 0, len(allSpending))
	for i := range allSpending {
		// Avoid funky pointer issues with arrays and for loops.
		spending := allSpending[i]

		// Skip spending objects that are not stale, or ones that are paused.
		if !spending.GetIsStale(now) || spending.GetIsPaused() {
			continue
		}

		fundingSchedule, ok := fundingSchedules[spending.FundingScheduleId]
		if !ok {
			fundingSchedule, err = p.repo.GetFundingSchedule(
				span.Context(),
				spending.BankAccountId,
				spending.FundingScheduleId,
			)
			if err != nil {
				log.WarnContext(span.Context(), "failed to retrieve funding schedule for spending object, it will not be processed", "err", err)
				continue
			}

			fundingSchedules[spending.FundingScheduleId] = fundingSchedule
		}

		spending.CalculateNextContribution(
			span.Context(),
			timezone,
			fundingSchedule,
			now,
			log,
		)

		spendingToUpdate = append(spendingToUpdate, spending)
	}

	if len(spendingToUpdate) == 0 {
		log.InfoContext(span.Context(), "no stale spending object were updated")
		return nil
	}

	log.InfoContext(span.Context(), "updating stale spending objects", "count", len(spendingToUpdate))

	return errors.Wrap(p.repo.UpdateSpending(
		span.Context(),
		p.args.BankAccountId,
		spendingToUpdate,
	), "failed to update stale spending")
}

func ProcessSpendingCron(ctx queue.Context) error {
	log := ctx.Log()

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())
	bankAccountsWithStaleSpending, err := jobRepo.GetBankAccountsWithStaleSpending(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve bank accounts with stale spending")
	}

	if len(bankAccountsWithStaleSpending) == 0 {
		crumbs.Debug(ctx, "No bank accounts had stale spending objects.", nil)
		log.InfoContext(ctx, "no bank accounts have stale spending objects")
		return nil
	}

	log.InfoContext(ctx, "found bank accounts with stale spending", "count", len(bankAccountsWithStaleSpending))
	crumbs.Debug(ctx, "Found bank accounts with stale spending.", map[string]any{
		"count": len(bankAccountsWithStaleSpending),
	})

	jobErrors := make([]error, 0)

	for _, item := range bankAccountsWithStaleSpending {
		itemLog := log.With(
			"accountId", item.AccountId,
			"bankAccountId", item.BankAccountId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing bank account to process stale spending")

		err = queue.Enqueue(
			ctx,
			ctx.Processor(),
			ProcessSpending,
			ProcessSpendingArguments{
				AccountId:     item.AccountId,
				BankAccountId: item.BankAccountId,
			},
		)
		if err != nil {
			log.WarnContext(ctx, "failed to enqueue job to process stale spending", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to process stale spending", "job", map[string]any{
				"error": err,
			})
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued bank accounts for stale spending processing")
	}

	return nil
}

func ProcessSpending(ctx queue.Context, args ProcessSpendingArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"bankAccountId", args.BankAccountId,
		)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
			args.AccountId,
			ctx.DB(),
			log,
		)

		account, err := repo.GetAccount(ctx)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve account to process stale spending", "err", err)
			return err
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.ErrorContext(ctx, "failed to parse account timezone", "err", err)
			return err
		}

		now := ctx.Clock().Now()
		allSpending, err := repo.GetSpending(ctx, args.BankAccountId)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve spending for bank account", "err", err)
			return err
		}

		fundingSchedules := map[ID[FundingSchedule]]*FundingSchedule{}

		spendingToUpdate := make([]Spending, 0, len(allSpending))
		for i := range allSpending {
			// Avoid funky pointer issues with arrays and for loops.
			spending := allSpending[i]

			// Skip spending objects that are not stale, or ones that are paused.
			if !spending.GetIsStale(now) || spending.GetIsPaused() {
				continue
			}

			fundingSchedule, ok := fundingSchedules[spending.FundingScheduleId]
			if !ok {
				fundingSchedule, err = repo.GetFundingSchedule(
					ctx,
					spending.BankAccountId,
					spending.FundingScheduleId,
				)
				if err != nil {
					log.WarnContext(ctx, "failed to retrieve funding schedule for spending object, it will not be processed", "err", err)
					continue
				}

				fundingSchedules[spending.FundingScheduleId] = fundingSchedule
			}

			spending.CalculateNextContribution(
				ctx,
				timezone,
				fundingSchedule,
				now,
				log,
			)

			spendingToUpdate = append(spendingToUpdate, spending)
		}

		if len(spendingToUpdate) == 0 {
			log.InfoContext(ctx, "no stale spending object were updated")
			return nil
		}

		log.InfoContext(ctx, "updating stale spending objects", "count", len(spendingToUpdate))

		return errors.Wrap(repo.UpdateSpending(
			ctx,
			args.BankAccountId,
			spendingToUpdate,
		), "failed to update stale spending")
	})
}
