import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { ArrowUpDown, HeartCrack, PiggyBank, Save, Trash } from 'lucide-react';
import { useLocation, useParams } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormCheckbox from '@monetr/interface/components/FormCheckbox';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import GoalTimeline from '@monetr/interface/components/goals/GoalTimeline';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
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

interface GoalValues {
  name: string;
  amount: number;
  nextRecurrence: Date;
  fundingScheduleId: ID<FundingSchedule>;
  isPaused: boolean;
}

export default function GoalDetails(): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const removeSpending = useRemoveSpending();
  const updateSpending = useUpdateSpending();
  const [, navigate] = useLocation();
  const { spendingId } = useParams<{ spendingId: ID<Spending> }>();
  const { enqueueSnackbar } = useSnackbar();
  const { data: spending, isLoading, isError } = useSpending(spendingId);

  if (!spendingId) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>There wasn&apos;t a goal specified...</Typography>
      </div>
    );
  }

  // Treat the locale still loading the same as the spending still loading, otherwise we fall all the way through to the
  // null return below and flash a blank page while the currency formatting catches up.
  if (isLoading || !locale) {
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
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>Couldn&apos;t find the goal you specified...</Typography>
      </div>
    );
  }

  if (!spending) {
    return null;
  }

  if (spending.spendingType !== SpendingType.Goal) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>This spending object is not a goal...</Typography>
      </div>
    );
  }

  function backToGoals() {
    navigate(`/bank/${spending?.bankAccountId}/goals`);
  }

  async function deleteGoal(): Promise<void> {
    if (!spending) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete goal: ${spending.name}`)) {
      return removeSpending(spending.spendingId).then(() => backToGoals());
    }

    return Promise.resolve();
  }

  async function submit(values: GoalValues, helpers: FormikHelpers<GoalValues>): Promise<void> {
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
      ruleset: null,
      targetAmount: locale.friendlyToAmount(values.amount),
      isPaused: values.isPaused,
    });

    return updateSpending(updatedSpending)
      .then(
        () =>
          void enqueueSnackbar('Updated goal successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error || 'Failed to update goal', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: GoalValues = {
    name: spending.name,
    amount: locale.amountToFriendly(spending.targetAmount),
    nextRecurrence: spending.nextRecurrence ?? startOfTomorrow({ in: inTimezone }),
    fundingScheduleId: spending.fundingScheduleId,
    isPaused: spending.isPaused,
  };

  const usedProgress = (
    (Math.min(spending?.usedAmount, spending?.targetAmount) / spending?.targetAmount) *
    100
  ).toFixed(0);
  const allocatedProgress = (
    (Math.min(spending?.currentAmount + spending?.usedAmount, spending?.targetAmount) / spending?.targetAmount) *
    100
  ).toFixed(0);

  return (
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
      <MTopNavigation
        base={`/bank/${spending.bankAccountId}/goals`}
        breadcrumb={spending?.name}
        icon={PiggyBank}
        title='Goals'
      >
        <Button onClick={() => showTransferModal({ initialToSpendingId: spending?.spendingId })} variant='secondary'>
          <ArrowUpDown />
          Transfer
        </Button>
        <Button onClick={deleteGoal} variant='destructive'>
          <Trash />
          Remove
        </Button>
        <FormButton role='form' type='submit' variant='primary'>
          <Save />
          Save
        </FormButton>
      </MTopNavigation>
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
                <div className={styles.progressFillAllocated} style={{ width: `${allocatedProgress}%` }} />
                <div className={styles.progressFillUsed} style={{ width: `${usedProgress}%` }} />
              </div>
            </div>

            <Divider className={styles.dividerHalf} />

            <FormTextField
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
              label='Target Date'
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
            <FormCheckbox
              data-testid='goal-details-paused'
              description='Pause this goal to temporarily stop contributions to it.'
              label='Paused?'
              name='isPaused'
            />
          </div>
          <Divider className={styles.dividerMobile} />
          <div className={styles.columnTimeline}>
            <Typography className={styles.timelineTitle} size='xl'>
              Goal Timeline
            </Typography>
            <GoalTimeline spendingId={spending.spendingId} />
          </div>
        </div>
      </div>
    </MForm>
  );
}
