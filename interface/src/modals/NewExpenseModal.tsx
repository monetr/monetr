import { useId, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { isBefore, startOfDay, startOfTomorrow } from 'date-fns';
import { type FormikHelpers, useFormikContext } from 'formik';
import { Repeat } from 'lucide-react';
import { rrulestr } from 'rrule';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import { Switch } from '@monetr/interface/components/Switch';
import Typography from '@monetr/interface/components/Typography';
import { useCreateSpending } from '@monetr/interface/hooks/useCreateSpending';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency, { type LocaleCurrency } from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchTransaction } from '@monetr/interface/hooks/usePatchTransaction';
import {
  type RecurringPriceChange,
  useRecurringTransactionHistory,
} from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import { SpendingType } from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './NewExpenseModal.module.scss';

interface NewExpenseValues {
  name: string;
  amount: number;
  nextOccurrence: Date;
  ruleset: string;
  fundingScheduleId: ID<FundingSchedule>;
  autoCreateTransaction: boolean;
  moveTransaction: boolean;
}

export interface NewExpenseModalProps {
  /**
   * recurring fills in the new expense from a recurring transaction and links the expense to it once its created
   */
  recurring?: TransactionRecurring;
  /**
   * transaction is the one the user started from when creating an expense from a recurring transaction. its name is the
   * fallback if theres no cluster name, and it lets them spend that charge from the new expense
   */
  transaction?: Transaction;
}

function NewExpenseModal(props: NewExpenseModalProps): React.JSX.Element {
  const { recurring, transaction } = props;
  const { timezone, inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const { data: dateLocale } = useLocale();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const { data: link } = useCurrentLink();
  const isManual = Boolean(link?.getIsManual());
  const createSpending = useCreateSpending();
  const patchTransaction = usePatchTransaction();
  const { seen, priceChange } = useRecurringTransactionHistory(recurring);
  // use the similar transactions name for the expense instead of the transactions own name, since thats the name for the
  // whole group of charges and not just the one they happened to start from
  const { data: cluster, isLoading: clusterIsLoading } = useTransactionCluster(recurring?.transactionClusterId ?? null);
  const name = cluster?.name || transaction?.getName();

  const ref = useRef<MModalRef>(null);

  // wait for the cluster too, the form only reads its initial values once so the name has to be there up front
  if (!selectedBankAccount || !locale || clusterIsLoading) {
    return (
      <MModal className={styles.modal} open={modal.visible} ref={ref}>
        One moment...
      </MModal>
    );
  }

  const tomorrow = startOfTomorrow({
    in: inTimezone,
  });
  const initialValues: NewExpenseValues = {
    name: name ?? '',
    amount: recurring ? locale.amountToFriendly(Math.abs(recurring.lastAmount)) : 0.0,
    nextOccurrence: recurring ? getNextRecurrence(recurring, tomorrow) : tomorrow,
    ruleset: recurring?.ruleset ?? '',
    fundingScheduleId: ID.from<FundingSchedule, string>(''),
    autoCreateTransaction: false,
    moveTransaction: false,
  };

  async function submit(values: NewExpenseValues, helper: FormikHelpers<NewExpenseValues>): Promise<void> {
    if (!selectedBankAccount || !locale) {
      return Promise.resolve();
    }

    helper.setSubmitting(true);
    return await createSpending({
      bankAccountId: selectedBankAccount.bankAccountId,
      transactionRecurringId: recurring?.transactionRecurringId ?? null,
      name: values.name.trim(),
      nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
        in: inTimezone,
      }),
      spendingType: SpendingType.Expense,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: locale.friendlyToAmount(values.amount),
      ruleset: values.ruleset,
      // Auto create transaction requires a manual link and a non-zero target
      // amount; force it off otherwise so the API will not reject the create.
      autoCreateTransaction: isManual && values.amount > 0 && values.autoCreateTransaction,
    })
      .then(async created => {
        if (transaction && values.moveTransaction) {
          await patchTransaction({
            transactionId: transaction.transactionId,
            bankAccountId: transaction.bankAccountId,
            spendingId: created.spendingId,
          });
        }
        return created;
      })
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} data-testid='new-expense-modal' initialValues={initialValues} onSubmit={submit}>
        <div className={styles.body}>
          <Typography className={styles.heading} size='xl' weight='bold'>
            Create A New Expense
          </Typography>
          {recurring && (
            <div className={styles.recurringBanner} data-testid='new-expense-recurring-banner'>
              <Repeat />
              <Typography color='inherit' size='sm'>
                Filled in from your {seen} {name ?? 'recurring'} charges. Give it a once over before you create it.
              </Typography>
            </div>
          )}
          <FormTextField
            autoComplete='off'
            autoFocus
            data-1p-ignore
            label='What are you budgeting for?'
            name='name'
            placeholder='Amazon, Netflix...'
            required
          />
          <div className={styles.fieldRow}>
            <FormAmountField
              allowNegative={false}
              className={styles.fieldRowItem}
              description={recurring && getAmountDescription(priceChange, locale, dateLocale?.code, timezone)}
              label='How much do you need?'
              name='amount'
              required
            />
            <FormDatePicker
              className={styles.fieldRowItem}
              description={recurring && 'When the next charge is expected.'}
              label='When do you need it next?'
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='nextOccurrence'
              required
            />
          </div>
          <MSelectFunding
            label='When do you want to fund the expense?'
            menuPortalTarget={document.body}
            name='fundingScheduleId'
            required
          />
          <MSelectFrequency
            dateFrom='nextOccurrence'
            description={recurring && `Matches when ${name ?? 'it'} charges you.`}
            label='How frequently do you need this expense?'
            name='ruleset'
            placeholder='Select a spending frequency...'
            required
          />
          {isManual && <AutoCreateTransactionToggle />}
          {transaction && !transaction.getIsAddition() && <MoveTransactionToggle transaction={transaction} />}
        </div>
        <div className={styles.actions}>
          <Button data-testid='close-new-expense-modal' onClick={modal.remove} variant='secondary'>
            Cancel
          </Button>
          <FormButton type='submit' variant='primary'>
            Create
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

function AutoCreateTransactionToggle(): React.JSX.Element {
  const autoCreateSwitchId = useId();
  const { setFieldValue, values } = useFormikContext<NewExpenseValues>();
  const hasAmount = (values.amount ?? 0) > 0;

  return (
    <div className={styles.optionRow} data-testid='new-expense-auto-create-transaction'>
      <div className={styles.optionText}>
        <label aria-disabled={!hasAmount} className={styles.optionLabel} htmlFor={autoCreateSwitchId}>
          Auto create transaction
        </label>
        <p aria-disabled={!hasAmount} className={styles.optionDescription}>
          Automatically add a transaction for this expense each time it is due, deducting from your balance.
        </p>
      </div>
      <Switch
        checked={hasAmount && values.autoCreateTransaction}
        disabled={!hasAmount}
        id={autoCreateSwitchId}
        onCheckedChange={() => setFieldValue('autoCreateTransaction', !values.autoCreateTransaction)}
      />
    </div>
  );
}

interface MoveTransactionToggleProps {
  transaction: Transaction;
}

function MoveTransactionToggle({ transaction }: MoveTransactionToggleProps): React.JSX.Element {
  const moveSwitchId = useId();
  const { setFieldValue, values } = useFormikContext<NewExpenseValues>();
  const { timezone } = useTimezone();
  const { data: locale } = useLocale();
  const date = locale
    ? new Intl.DateTimeFormat(locale.code, { month: 'short', day: 'numeric', timeZone: timezone }).format(
        transaction.date,
      )
    : null;

  return (
    <div className={styles.optionRow} data-testid='new-expense-move-transaction'>
      <div className={styles.optionText}>
        <label className={styles.optionLabel} htmlFor={moveSwitchId}>
          Spend the {date} charge from this expense
        </label>
        <p className={styles.optionDescription}>
          This won&apos;t change your Free-To-Use balance, since the expense is brand new.
        </p>
      </div>
      <Switch
        checked={values.moveTransaction}
        id={moveSwitchId}
        onCheckedChange={() => setFieldValue('moveTransaction', !values.moveTransaction)}
      />
    </div>
  );
}

// getNextRecurrence returns when the recurring transaction is next expected. next only gets updated when its recalculated
// though so it can already be in the past, in that case use the next occurrence of the rule instead since an expense
// cant be due in the past
function getNextRecurrence(recurring: TransactionRecurring, tomorrow: Date): Date {
  if (!isBefore(recurring.next, tomorrow)) {
    return recurring.next;
  }

  return rrulestr(recurring.ruleset).after(tomorrow, true) ?? tomorrow;
}

function getAmountDescription(
  priceChange: RecurringPriceChange | null,
  locale: LocaleCurrency,
  localeCode: string | undefined,
  timezone: string,
): string {
  if (!priceChange) {
    return 'Your last charge.';
  }

  const previous = locale.formatAmount(Math.abs(priceChange.previousAmount), AmountType.Stored);
  const month = new Intl.DateTimeFormat(localeCode, { month: 'long', timeZone: timezone }).format(
    priceChange.changedAt,
  );
  return `Your last charge. It was ${previous} until ${month}.`;
}

const newExpenseModal = NiceModal.create(NewExpenseModal);

export default newExpenseModal;

export function showNewExpenseModal(props: NewExpenseModalProps = {}): Promise<Spending | null> {
  return NiceModal.show(newExpenseModal, props) as Promise<Spending | null>;
}
