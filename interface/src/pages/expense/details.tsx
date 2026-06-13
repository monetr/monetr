import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { ArrowUpDown, HeartCrack, Receipt, Save, Trash } from 'lucide-react';
import { useLocation, useParams } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import ExpenseTransactionList from '@monetr/interface/components/expenses/ExpenseTransactionList';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormCheckbox from '@monetr/interface/components/FormCheckbox';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRemoveSpending } from '@monetr/interface/hooks/useRemoveSpending';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useUpdateSpending } from '@monetr/interface/hooks/useUpdateSpending';
import { showTransferModal } from '@monetr/interface/modals/TransferModal';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { ID } from '@monetr/interface/models/ID';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './details.module.scss';
import ExpenseTimeline from './ExpenseTimeline';

interface ExpenseValues {
  name: string;
  amount: number;
  nextRecurrence: Date;
  fundingScheduleId: ID<FundingSchedule>;
  ruleset: string;
  autoCreateTransaction: boolean;
}

export default function ExpenseDetails(): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const removeSpending = useRemoveSpending();
  const updateSpending = useUpdateSpending();
  const [, navigate] = useLocation();
  const { spendingId } = useParams<{ spendingId: ID<Spending> }>();
  const { enqueueSnackbar } = useSnackbar();
  const { data: spending, isLoading, isError } = useSpending(spendingId);
  const { data: link } = useCurrentLink();
  const isManual = Boolean(link?.getIsManual());

  if (!spendingId) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>There wasn't an expense specified...</Typography>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>Couldn't find the expense you specified...</Typography>
      </div>
    );
  }

  if (!spending || !locale) {
    return null;
  }

  if (spending.spendingType !== SpendingType.Expense) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>This spending object is not an expense...</Typography>
      </div>
    );
  }

  function backToExpenses() {
    navigate(`/bank/${spending?.bankAccountId}/expenses`);
  }

  async function deleteExpense(): Promise<void> {
    if (!spending) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete expense: ${spending.name}`)) {
      return removeSpending(spending.spendingId).then(() => backToExpenses());
    }

    return Promise.resolve();
  }

  async function submit(values: ExpenseValues, helpers: FormikHelpers<ExpenseValues>): Promise<void> {
    if (!spending || !locale) {
      return Promise.resolve();
    }

    helpers.setSubmitting(true);

    const updatedSpending = new Spending({
      ...spending,
      name: values.name,
      nextRecurrence: startOfDay(values.nextRecurrence, {
        in: inTimezone,
      }),
      fundingScheduleId: values.fundingScheduleId,
      ruleset: values.ruleset,
      targetAmount: locale.friendlyToAmount(values.amount),
      // Auto create transaction is only supported on manual links; force it off
      // otherwise so the API will not reject the update.
      autoCreateTransaction: isManual && values.autoCreateTransaction,
    });

    return updateSpending(updatedSpending)
      .then(
        () =>
          void enqueueSnackbar('Updated expense successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error || 'Failed to update expense', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: ExpenseValues = {
    name: spending.name,
    amount: locale.amountToFriendly(spending.targetAmount),
    nextRecurrence: spending.nextRecurrence ?? startOfTomorrow({ in: inTimezone }),
    fundingScheduleId: spending.fundingScheduleId,
    ruleset: spending.ruleset ?? '',
    autoCreateTransaction: spending.autoCreateTransaction,
  };

  const progress = ((Math.min(spending?.currentAmount, spending?.targetAmount) / spending?.targetAmount) * 100).toFixed(
    0,
  );

  return (
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
      <MTopNavigation
        base={`/bank/${spending.bankAccountId}/expenses`}
        breadcrumb={spending?.name}
        icon={Receipt}
        title='Expenses'
      />
      <div className={styles.body}>
        <div className={styles.columns}>
          <div className={styles.column}>
            <div className={styles.summary}>
              <div className={styles.summaryHeader}>
                <MerchantIcon className={styles.flexNone} name={spending?.name} />
                <div className={styles.summaryText}>
                  <p className={styles.summaryName}>{spending?.name}</p>
                  <Typography size='inherit' weight='semibold'>
                    {locale.formatAmount(spending?.currentAmount, AmountType.Stored)}
                    <span className={styles.ofText}>of</span>
                    {locale.formatAmount(spending?.targetAmount, AmountType.Stored)}
                  </Typography>
                </div>
              </div>
              <div className={styles.progressTrack}>
                <div className={styles.progressFill} style={{ width: `${progress}%` }} />
              </div>
            </div>

            <Divider className={styles.dividerHalf} />

            <FormTextField
              autoComplete='off'
              className={layoutVariants({ width: 'full' })}
              data-1p-ignore
              label='Expense'
              name='name'
              required
            />
            <FormAmountField
              allowNegative={false}
              className={layoutVariants({ width: 'full' })}
              label='Amount'
              name='amount'
              required
            />
            <FormDatePicker
              className={layoutVariants({ width: 'full' })}
              label='Next Occurrence'
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='nextRecurrence'
              required
            />
            <MSelectFunding
              className={layoutVariants({ width: 'full' })}
              label='When do you want to fund the expense?'
              menuPortalTarget={document.body}
              name='fundingScheduleId'
              required
            />
            <MSelectFrequency
              className={layoutVariants({ width: 'full' })}
              dateFrom='nextRecurrence'
              label='How often do you need this expense?'
              name='ruleset'
              placeholder='Select a spending frequency...'
              required
            />
            {isManual && (
              <FormCheckbox
                className={layoutVariants({ width: 'full' })}
                description='Automatically add a transaction for this expense each time it is due, deducting from your balance.'
                label='Auto create transaction'
                name='autoCreateTransaction'
              />
            )}
            <div className={styles.formButtons}>
              <Button
                onClick={() => showTransferModal({ initialToSpendingId: spending?.spendingId })}
                variant='secondary'
              >
                <ArrowUpDown />
                Transfer
              </Button>
              <Button onClick={deleteExpense} variant='destructive'>
                <Trash />
                Remove
              </Button>
              <FormButton role='form' type='submit' variant='primary'>
                <Save />
                Save
              </FormButton>
            </div>
            <ExpenseTransactionList spending={spending} />
          </div>
          <Divider className={styles.dividerMobile} />
          <div className={styles.columnTimeline}>
            <Typography className={styles.timelineTitle} size='xl'>
              Expense Timeline
            </Typography>
            <ExpenseTimeline spendingId={spending.spendingId} />
          </div>
        </div>
      </div>
    </MForm>
  );
}
