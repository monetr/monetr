package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ProcessSpending = "ProcessSpending"
)

type (
	ProcessSpendingHandler struct {
		log          *logrus.Entry
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
		log   *logrus.Entry
		repo  repository.BaseRepository
		clock clock.Clock
	}
)

func NewProcessSpendingHandler(
	log *logrus.Entry,
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
	return ProcessSpending
}

func (p *ProcessSpendingHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args ProcessSpendingArguments
	if err := errors.Wrap(p.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Process Spending job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log = log.WithContext(span.Context()).WithFields(logrus.Fields{
			"accountId":     args.AccountId,
			"bankAccountId": args.BankAccountId,
		})

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
	log := p.log.WithContext(ctx)

	log.Info("retrieving bank accounts with stale spending")
	bankAccountsWithStaleSpending, err := p.repo.GetBankAccountsWithStaleSpending(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve bank accounts with stale spending")
	}

	if len(bankAccountsWithStaleSpending) == 0 {
		crumbs.Debug(ctx, "No bank accounts had stale spending objects.", nil)
		log.Info("no bank accounts have stale spending objects")
		return nil
	}

	log.WithField("count", len(bankAccountsWithStaleSpending)).Info("found bank accounts with stale spending")
	crumbs.Debug(ctx, "Found bank accounts with stale spending.", map[string]interface{}{
		"count": len(bankAccountsWithStaleSpending),
	})

	jobErrors := make([]error, 0)

	for _, item := range bankAccountsWithStaleSpending {
		itemLog := log.WithFields(logrus.Fields{
			"accountId":     item.AccountId,
			"bankAccountId": item.BankAccountId,
		})
		itemLog.Trace("enqueuing bank account to process stale spending")
		err = enqueuer.EnqueueJob(ctx, p.QueueName(), ProcessSpendingArguments{
			AccountId:     item.AccountId,
			BankAccountId: item.BankAccountId,
		})
		if err != nil {
			log.WithError(err).Warn("failed to enqueue job to process stale spending")
			crumbs.Warn(ctx, "Failed to enqueue job to process stale spending", "job", map[string]interface{}{
				"error": err,
			})
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Trace("successfully enqueued bank accounts for stale spending processing")
	}

	return nil
}

func NewProcessSpendingJob(
	log *logrus.Entry,
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

	log := p.log.WithContext(span.Context())

	account, err := p.repo.GetAccount(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve account to process stale spending")
		return err
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		log.WithError(err).Error("failed to parse account timezone")
		return err
	}

	now := p.clock.Now()
	allSpending, err := p.repo.GetSpending(span.Context(), p.args.BankAccountId)
	if err != nil {
		log.WithError(err).Error("failed to retrieve spending for bank account")
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
				log.WithError(err).Warn("failed to retrieve funding schedule for spending object, it will not be processed")
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
		log.Info("no stale spending object were updated")
		return nil
	}

	log.WithField("count", len(spendingToUpdate)).Info("updating stale spending objects")

	return errors.Wrap(p.repo.UpdateSpending(
		span.Context(),
		p.args.BankAccountId,
		spendingToUpdate,
	), "failed to update stale spending")
}
