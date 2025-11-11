import type { AxiosError } from 'axios';
import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { ArrowUpDown, HeartCrack, PiggyBank, Save, Trash } from 'lucide-react';
import { useSnackbar } from 'notistack';
import { useNavigate, useParams } from 'react-router-dom';

import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormCheckbox from '@monetr/interface/components/FormCheckbox';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import GoalTimeline from '@monetr/interface/components/goals/GoalTimeline';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRemoveSpending } from '@monetr/interface/hooks/useRemoveSpending';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useUpdateSpending } from '@monetr/interface/hooks/useUpdateSpending';
import { showTransferModal } from '@monetr/interface/modals/TransferModal';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import type { APIError } from '@monetr/interface/util/request';

interface GoalValues {
  name: string;
  amount: number;
  nextRecurrence: Date;
  fundingScheduleId: string;
  isPaused: boolean;
}

export default function GoalDetails(): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const removeSpending = useRemoveSpending();
  const updateSpending = useUpdateSpending();
  const navigate = useNavigate();
  const { spendingId } = useParams();
  const { enqueueSnackbar } = useSnackbar();
  const { data: spending, isLoading, isError } = useSpending(spendingId);

  if (!spendingId) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content size-24' />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>There wasn't a goal specified...</Typography>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content size-24' />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>Couldn't find the goal you specified...</Typography>
      </div>
    );
  }

  if (!spending) {
    return null;
  }

  if (spending.spendingType !== SpendingType.Goal) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content size-24' />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>This spending object is not a goal...</Typography>
      </div>
    );
  }

  function backToGoals() {
    navigate(`/bank/${spending.bankAccountId}/goals`);
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
    helpers.setSubmitting(true);

    const updatedSpending = new Spending({
      ...spending,
      name: values.name,
      description: null,
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
        (error: AxiosError<APIError>) =>
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
    nextRecurrence: spending.nextRecurrence,
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
    <MForm initialValues={initialValues} onSubmit={submit} className='flex w-full h-full flex-col'>
      <MTopNavigation
        icon={PiggyBank}
        title='Goals'
        base={`/bank/${spending.bankAccountId}/goals`}
        breadcrumb={spending?.name}
      >
        <Button variant='secondary' onClick={() => showTransferModal({ initialToSpendingId: spending?.spendingId })}>
          <ArrowUpDown />
          Transfer
        </Button>
        <Button variant='destructive' onClick={deleteGoal}>
          <Trash />
          Remove
        </Button>
        <FormButton variant='primary' type='submit' role='form'>
          <Save />
          Save
        </FormButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4 pb-16 md:pb-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col'>
            <div className='flex flex-col w-full'>
              <div className='flex gap-4 items-center w-full overflow-hidden'>
                <MerchantIcon name={spending?.name} className='flex-none' />
                <div className='flex flex-col flex-1 overflow-hidden'>
                  <p className='text-ellipsis truncate min-w-0'>{spending?.name}</p>
                  <MSpan weight='semibold'>
                    {locale.formatAmount(spending?.currentAmount, AmountType.Stored)}
                    <span className='font-normal'>of</span>
                    {locale.formatAmount(spending?.targetAmount, AmountType.Stored)}
                  </MSpan>
                </div>
              </div>
              <div className='w-full bg-gray-200 rounded-full h-1.5 my-2 dark:bg-gray-700 relative'>
                <div
                  className='absolute top-0 bg-green-600 h-1.5 rounded-full dark:bg-green-600'
                  style={{ width: `${allocatedProgress}%` }}
                />
                <div
                  className='absolute top-0 bg-blue-600 h-1.5 rounded-full dark:bg-blue-600'
                  style={{ width: `${usedProgress}%` }}
                />
              </div>
            </div>

            <Divider className='w-1/2 my-4' />

            <FormTextField className='w-full' label='Expense' name='name' required data-1p-ignore />
            <FormAmountField allowNegative={false} className='w-full' label='Amount' name='amount' required />
            <FormDatePicker
              label='Target Date'
              name='nextRecurrence'
              className='w-full'
              required
              min={startOfTomorrow({
                in: inTimezone,
              })}
            />
            <MSelectFunding
              className='w-full'
              label='When do you want to fund the expense?'
              menuPortalTarget={document.body}
              name='fundingScheduleId'
              required
            />
            <FormCheckbox
              data-testid='goal-details-paused'
              name='isPaused'
              label='Paused?'
              description='Pause this goal to temporarily stop contributions to it.'
            />
          </div>
          <Divider className='block md:hidden w-1/2' />
          <div className='w-full md:w-1/2 flex flex-col gap-2'>
            <MSpan className='text-xl my-2'>Goal Timeline</MSpan>
            <GoalTimeline spendingId={spending.spendingId} />
          </div>
        </div>
      </div>
    </MForm>
  );
}
