package background

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/elliotcourant/gofx"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/formats/ofx"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ProcessOFXUpload = "ProcessOFXUpload"
)

var (
	_ JobHandler        = &ProcessOFXUploadHandler{}
	_ JobImplementation = &ProcessOFXUploadJob{}
)

type (
	ProcessOFXUploadHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		publisher    pubsub.Publisher
		files        storage.Storage
		enqueuer     JobEnqueuer
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	ProcessOFXUploadArguments struct {
		AccountId           ID[Account]           `json:"accountId"`
		BankAccountId       ID[BankAccount]       `json:"bankAccountId"`
		TransactionUploadId ID[TransactionUpload] `json:"transactionUploadId"`
	}

	ProcessOFXUploadJob struct {
		args      ProcessOFXUploadArguments
		log       *logrus.Entry
		repo      repository.BaseRepository
		files     storage.Storage
		publisher pubsub.Publisher
		enqueuer  JobEnqueuer
		clock     clock.Clock
		timezone  *time.Location

		upload                *TransactionUpload
		file                  *File
		data                  *gofx.OFX
		currency              string
		statementTransactions []gofx.StatementTransaction
		existingTransactions  map[string]Transaction
	}
)

func NewProcessOFXUploadHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	files storage.Storage,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *ProcessOFXUploadHandler {
	return &ProcessOFXUploadHandler{
		log:          log,
		db:           db,
		publisher:    publisher,
		files:        files,
		enqueuer:     enqueuer,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (h *ProcessOFXUploadHandler) updateStatus(
	ctx context.Context,
	args ProcessOFXUploadArguments,
	status TransactionUploadStatus,
	errorMessage *string,
) error {
	log := h.log.WithContext(ctx).WithFields(logrus.Fields{
		"accountId":           args.AccountId,
		"bankAccountId":       args.BankAccountId,
		"transactionUploadId": args.TransactionUploadId,
	})

	query := h.db.ModelContext(ctx, &TransactionUpload{}).
		Where(`"account_id" = ?`, args.AccountId).
		Where(`"bank_account_id" = ?`, args.BankAccountId).
		Where(`"transaction_upload_id" = ?`, args.TransactionUploadId).
		Set(`"status" = ?`, status)

	switch status {
	case TransactionUploadStatusProcessing:
		query = query.Set(`"processed_at" = ?`, h.clock.Now())
	case TransactionUploadStatusComplete:
		query = query.Set(`"completed_at" = ?`, h.clock.Now())
	case TransactionUploadStatusFailed:
		query = query.Set(`"completed_at" = ?`, h.clock.Now())
		if errorMessage != nil {
			query = query.Set(`"error" = ?`, *errorMessage)
		} else {
			query = query.Set(`"error" = ?`, "Unknown failure")
		}
	}

	log.WithField("status", status).Trace("updated transaction upload status")

	_, err := query.Update()
	if err != nil {
		return errors.Wrap(err, "failed to update upload status")
	}

	channel := fmt.Sprintf(
		"account:%s:transaction_upload:%s:progress",
		args.AccountId, args.TransactionUploadId,
	)
	payload := string(status)
	if err := h.publisher.Notify(ctx, channel, payload); err != nil {
		return errors.Wrap(err, "failed to send progress notification for job")
	}
	log.WithFields(logrus.Fields{
		"channel": channel,
		"payload": payload,
	}).Trace("sent progress notification for file upload")

	return nil
}

func (h *ProcessOFXUploadHandler) QueueName() string {
	return ProcessOFXUpload
}

func (h *ProcessOFXUploadHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args ProcessOFXUploadArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Processing OFX Upload job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	if err := h.updateStatus(ctx, args, TransactionUploadStatusProcessing, nil); err != nil {
		return err
	}

	// error here then we will catch it and update the upload status accordingly.
	var err error
	defer func() {
		if recovery := recover(); recovery != nil {
			log.WithError(err).Error("panic processing OFX file upload")
			_ = h.updateStatus(ctx, args, TransactionUploadStatusFailed, nil)

			panic(recovery)
		}
		if err != nil {
			log.WithError(err).Error("error processing OFX file upload")
			errorString := fmt.Sprintf("%s", err)
			_ = h.updateStatus(ctx, args, TransactionUploadStatusFailed, &errorString)
		} else {
			_ = h.updateStatus(ctx, args, TransactionUploadStatusComplete, nil)
		}
	}()
	err = h.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context())
		repo := repository.NewRepositoryFromSession(h.clock, "user_system", args.AccountId, txn)

		job, err := NewProcessOFXUploadJob(
			log, repo, h.clock, h.files, h.publisher, h.enqueuer, args,
		)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})

	// Return the error anyway so we can see failed uploads in sentry.
	return err
}

func NewProcessOFXUploadJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	files storage.Storage,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
	args ProcessOFXUploadArguments,
) (*ProcessOFXUploadJob, error) {
	return &ProcessOFXUploadJob{
		args:      args,
		log:       log,
		repo:      repo,
		files:     files,
		publisher: publisher,
		enqueuer:  enqueuer,
		clock:     clock,
	}, nil
}

func (j *ProcessOFXUploadJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()
	crumbs.AddTag(span.Context(), "bankAccountId", j.args.BankAccountId.String())
	crumbs.AddTag(span.Context(), "transactionUploadId", j.args.TransactionUploadId.String())
	crumbs.IncludeUserInScope(span.Context(), j.args.AccountId)

	log := j.log.WithContext(span.Context())

	// No matter what, when we are finished clean up the file.
	defer func() {
		now := j.clock.Now()
		j.file.DeletedAt = &now
		log.Debug("processing complete, marking file as deleted and queueing removal")
		if err := j.repo.UpdateFile(span.Context(), j.file); err != nil {
			log.
				WithField("fileId", j.file.FileId).
				WithError(err).
				Warn("failed to update file with deleted at")
		}

		j.enqueuer.EnqueueJob(span.Context(), RemoveFile, RemoveFileArguments{
			AccountId: j.args.AccountId,
			FileId:    j.file.FileId,
		})
	}()

	account, err := j.repo.GetAccount(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve account for job")
		return err
	}

	j.timezone, err = account.GetTimezone()
	if err != nil {
		log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
		j.timezone = time.UTC
	}

	// Load the file and its data into memory.
	if err := j.loadFile(span.Context()); err != nil {
		return err
	}

	// Pull all of the transactions that already exist in our system from the file
	// so we can compare.
	if err := j.hydrateTransactions(span.Context()); err != nil {
		return err
	}

	// Push new and updated transactions to the database.
	if err := j.syncTransactions(span.Context()); err != nil {
		return err
	}

	if err := j.syncBalances(span.Context()); err != nil {
		return err
	}

	// Also kick off the transaction similarity job.
	j.enqueuer.EnqueueJob(span.Context(), CalculateTransactionClusters, CalculateTransactionClustersArguments{
		AccountId:     j.args.AccountId,
		BankAccountId: j.args.BankAccountId,
	})

	return nil
}

// loadFile will take the current arguments and load the data from the file
// upload itself into memory for processing.
func (j *ProcessOFXUploadJob) loadFile(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	txnUpload, err := j.repo.GetTransactionUpload(
		span.Context(),
		j.args.BankAccountId,
		j.args.TransactionUploadId,
	)
	if err != nil {
		return errors.Wrap(err, "failed to process OFX file upload")
	}
	j.upload = txnUpload

	file, err := j.repo.GetFile(span.Context(), txnUpload.FileId)
	if err != nil {
		return errors.Wrap(err, "could not get file for processing")
	}
	j.file = file

	if file.DeletedAt != nil {
		return errors.New("cannot import transactions from a deleted file")
	}

	fileReader, _, err := j.files.Read(span.Context(), file.BlobUri)
	if err != nil {
		return errors.Wrap(err, "failed to access file from storage")
	}
	defer fileReader.Close()

	ofxData, err := ofx.Parse(fileReader)
	if err != nil {
		return err
	}

	j.data = ofxData
	return nil
}

// hydrateTransactions takes all of the transactions that were present in the
// OFX file and tries to cross reference them with transactions that already
// exist in the database. It relies on the OFX file having a unique identifier
// for each transaction that is consistent between each download from the FI.
func (j *ProcessOFXUploadJob) hydrateTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	externalTransactionIds := make([]string, 0)

	// Gather the bank transactions
	if bankResponse := j.data.BANKMSGSRSV1; bankResponse != nil {
		for _, statementTransactions := range bankResponse.STMTTRNRS {
			for _, transaction := range statementTransactions.STMTRS.BANKTRANLIST.STMTTRN {
				// Someday we might need to also consider CORRECTFITID.
				externalTransactionIds = append(externalTransactionIds, transaction.FITID)
				j.statementTransactions = append(j.statementTransactions, *transaction)
			}
			if j.currency == "" {
				j.currency = strings.ToUpper(statementTransactions.STMTRS.CURDEF)
			}
		}
	} else if bankResponse := j.data.CREDITCARDMSGSRSV1; bankResponse != nil {
		for _, statementTransactions := range bankResponse.CCSTMTTRNRS {
			for _, transaction := range statementTransactions.CCSTMTRS.BANKTRANLIST.STMTTRN {
				// Someday we might need to also consider CORRECTFITID.
				externalTransactionIds = append(externalTransactionIds, transaction.FITID)
				j.statementTransactions = append(j.statementTransactions, *transaction)
			}
			if j.currency == "" {
				j.currency = strings.ToUpper(statementTransactions.CCSTMTRS.CURDEF)
			}
		}
	}

	// If we are unable to derive the currency code from the file itself then we
	// should fallback to monetr's default.
	if j.currency == "" {
		j.currency = consts.DefaultCurrencyCode
	}

	// Reverse the order of the arrray we store such that the order we insert the
	// transactions into the DB matches the order of the transactions in the
	// actual file.
	slices.Reverse(j.statementTransactions)

	// TODO Add others as needed. Not sure what other formats we'll end up seeing
	// over time.

	if len(externalTransactionIds) == 0 {
		return errors.Errorf("no external transaction IDs were found in the file, account type may not be supported")
	}

	var err error
	j.existingTransactions, err = j.repo.GetTransactonsByUploadIdentifier(
		span.Context(),
		j.args.BankAccountId,
		externalTransactionIds,
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve existing transactions for upload processing")
	}

	if count := len(j.existingTransactions); count > 0 {
		j.log.WithContext(span.Context()).WithFields(logrus.Fields{
			"existingTransactions": count,
		}).Debug("found existing transactions for upload")
	}

	return nil
}

func (j *ProcessOFXUploadJob) syncTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := j.log.WithContext(span.Context())

	transactionsToUpdate := make([]*Transaction, 0)
	transactionsToCreate := make([]Transaction, 0)
	for y := range j.statementTransactions {
		externalTransaction := j.statementTransactions[y]
		uploadIdentifier := externalTransaction.FITID
		tlog := log.WithFields(logrus.Fields{
			"uploadIdentifier": uploadIdentifier,
		})

		// Parse the amount in the specified currency.
		amount, err := currency.ParseFriendlyToAmount(
			externalTransaction.TRNAMT,
			j.currency,
		)
		if err != nil {
			tlog.WithError(err).
				WithField("trnamt", externalTransaction.TRNAMT).
				Error("failed to parse transaction amount")
			continue
		}
		// monetr uses negative amounts to represent deposits and positive to
		// represent debits. This is the opposite of the file format, so we need
		// to invert the amount.
		amount = amount * -1

		// Still need to figure out something to do with memo, but for now we can
		// take the name and trim it. Memo seems to behave a bit differently from
		// FI to FI. At NFCU for example it contains a larger more un-santized
		// version of the transaction name, but at US Bank it seems to contain
		// reference numbers that might be useful internally? But are definitely
		// not helpful here.
		name := strings.TrimSpace(externalTransaction.NAME)
		originalName := strings.TrimSpace(externalTransaction.MEMO)

		// Make sure that the original name and name are set. This way if name is
		// blank it will use the original name. And if original name is blank it
		// will use the name.
		name = myownsanity.CoalesceStrings(name, originalName)
		originalName = myownsanity.CoalesceStrings(originalName, name)

		// TODO Also parse DTAVAIL at some point
		date, err := ofx.ParseDate(externalTransaction.DTPOSTED, j.timezone)
		if err != nil {
			tlog.WithError(err).
				WithField("dtposted", externalTransaction.DTPOSTED).
				Error("failed to parse transaction date posted")
			continue
		}

		transaction, ok := j.existingTransactions[uploadIdentifier]
		if !ok {
			transaction = Transaction{
				TransactionId:        NewID(&transaction),
				AccountId:            j.args.AccountId,
				BankAccountId:        j.args.BankAccountId,
				Amount:               amount,
				Date:                 date,
				Name:                 name,
				OriginalName:         originalName,
				OriginalMerchantName: name,
				IsPending:            false, // OFX files don't show pending?
				UploadIdentifier:     &uploadIdentifier,
				Source:               TransactionSourceUpload,
			}
			transactionsToCreate = append(transactionsToCreate, transaction)
			continue
		}

		// TODO Process changes to an existing transaction.
	}

	// Persist any new transactions.
	if count := len(transactionsToCreate); count > 0 {
		log.WithField("new", count).Info("creating new transactions from import")
		if err := j.repo.InsertTransactions(span.Context(), transactionsToCreate); err != nil {
			return errors.Wrap(err, "failed to persist new transactions")
		}
	}

	// If there are any updated transactions persist those as well.
	if count := len(transactionsToUpdate); count > 0 {
		log.WithField("updated", count).Info("updating transactions from import")
		if err := j.repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
			return errors.Wrap(err, "failed to update transactions")
		}
	}

	return nil
}

func (j *ProcessOFXUploadJob) syncBalances(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// TODO Somehow keep track of the as of timestamp? This way if someone is
	// importing files out of order we could potentially avoid updating the
	// balance to an old value.
	var currentBalance, availableBalance, limitBalance int64
	var err error
	if j.data.BANKMSGSRSV1 != nil {
		for i := range j.data.BANKMSGSRSV1.STMTTRNRS {
			statementTransactions := j.data.BANKMSGSRSV1.STMTTRNRS[i]
			if statementTransactions.STMTRS.LEDGERBAL != nil {
				currentBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.STMTRS.LEDGERBAL.BALAMT,
					j.currency,
				)
				if err != nil {
					return errors.Wrap(err, "failed to parse ledger balance amount")
				}

			}

			if statementTransactions.STMTRS.AVAILBAL != nil {
				availableBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.STMTRS.AVAILBAL.BALAMT,
					j.currency,
				)
				if err != nil {
					return errors.Wrap(err, "failed to parse available balance amount")
				}
			}
		}
	} else if j.data.CREDITCARDMSGSRSV1 != nil {
		for i := range j.data.CREDITCARDMSGSRSV1.CCSTMTTRNRS {
			statementTransactions := j.data.CREDITCARDMSGSRSV1.CCSTMTTRNRS[i]
			if statementTransactions.CCSTMTRS.LEDGERBAL != nil {
				currentBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.CCSTMTRS.LEDGERBAL.BALAMT,
					j.currency,
				)
				if err != nil {
					return errors.Wrap(err, "failed to parse ledger balance amount")
				}
			}

			if statementTransactions.CCSTMTRS.AVAILBAL != nil {
				availableBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.CCSTMTRS.AVAILBAL.BALAMT,
					j.currency,
				)
				if err != nil {
					return errors.Wrap(err, "failed to parse available balance amount")
				}
			}
			// The limit for credit cards is equal to the amount currrently available
			// plus the inverse of any amount currently used.
			currentBalance = -1 * currentBalance
			limitBalance = availableBalance + currentBalance
		}
	}

	bankAccount, err := j.repo.GetBankAccount(span.Context(), j.args.BankAccountId)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve bank account for file import sync")
	}

	if j.currency != bankAccount.Currency {
		return errors.Errorf("Currency of OFX file does not match currency of bank account, file: [%s], account: [%s]", j.currency, bankAccount.Currency)
	}

	// TODO Log the previous value and the new one?
	bankAccount.CurrentBalance = currentBalance
	bankAccount.AvailableBalance = availableBalance
	bankAccount.LimitBalance = limitBalance

	if err := j.repo.UpdateBankAccount(span.Context(), bankAccount); err != nil {
		return errors.Wrap(err, "failed to update bank account balances")
	}

	return nil
}
