import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  DeleteOutlined,
  HeartBroken,
  SaveOutlined,
  SavingsOutlined,
  SwapVertOutlined,
} from '@mui/icons-material';
import { AxiosError } from 'axios';
import { startOfDay, startOfToday } from 'date-fns';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import GoalTimeline from '@monetr/interface/components/goals/GoalTimeline';
import MAmountField from '@monetr/interface/components/MAmountField';
import MFormButton, { MBaseButton } from '@monetr/interface/components/MButton';
import MCheckbox from '@monetr/interface/components/MCheckbox';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MDivider from '@monetr/interface/components/MDivider';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useRemoveSpending, useSpending, useUpdateSpending } from '@monetr/interface/hooks/spending';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { showTransferModal } from '@monetr/interface/modals/TransferModal';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import { APIError } from '@monetr/interface/util/request';

interface GoalValues {
  name: string;
  amount: number;
  nextRecurrence: Date;
  fundingScheduleId: string;
  isPaused: boolean;
}

export default function GoalDetails(): JSX.Element {
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
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          There wasn't a goal specified...
        </MSpan>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          Couldn't find the goal you specified...
        </MSpan>
      </div>
    );
  }

  if (!spending) {
    return null;
  }

  if (spending.spendingType !== SpendingType.Goal) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          This spending object is not a goal...
        </MSpan>
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

    if (window.confirm(`Are you sure you want to delete goal: ${ spending.name }`)) {
      return removeSpending(spending.spendingId)
        .then(() => backToGoals());
    }

    return Promise.resolve();
  }

  async function submit(values: GoalValues, helpers: FormikHelpers<GoalValues>): Promise<void> {
    helpers.setSubmitting(true);

    const updatedSpending = new Spending({
      ...spending,
      name: values.name,
      description: null,
      nextRecurrence: startOfDay(values.nextRecurrence),
      fundingScheduleId: values.fundingScheduleId,
      ruleset: null,
      targetAmount: locale.friendlyToAmount(values.amount),
      isPaused: values.isPaused,
    });

    return updateSpending(updatedSpending)
      .then(() => void enqueueSnackbar(
        'Updated goal successfully',
        {
          variant: 'success',
          disableWindowBlurListener: true,
        },
      ))
      .catch((error: AxiosError<APIError>) => void enqueueSnackbar(
        error.response.data.error || 'Failed to update goal',
        {
          variant: 'error',
          disableWindowBlurListener: true,
        },
      ))
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: GoalValues = {
    name: spending.name,
    amount: locale.amountToFriendly(spending.targetAmount),
    nextRecurrence: spending.nextRecurrence,
    fundingScheduleId: spending.fundingScheduleId,
    isPaused: spending.isPaused,
  };

  const usedProgress = ((Math.min(
    spending?.usedAmount,
    spending?.targetAmount,
  ) / spending?.targetAmount) * 100).toFixed(0);
  const allocatedProgress = ((Math.min(
    spending?.currentAmount + spending?.usedAmount,
    spending?.targetAmount,
  ) / spending?.targetAmount) * 100).toFixed(0);

  return (
    <MForm initialValues={ initialValues } onSubmit={ submit } className='flex w-full h-full flex-col'>
      <MTopNavigation
        icon={ SavingsOutlined }
        title='Goals'
        base={ `/bank/${spending.bankAccountId}/goals` }
        breadcrumb={ spending?.name }
      >
        <MBaseButton
          color='secondary'
          className='gap-1 py-1 px-2'
          onClick={ () => showTransferModal({ initialToSpendingId: spending?.spendingId }) }
        >
          <SwapVertOutlined />
          Transfer
        </MBaseButton>
        <MBaseButton color='cancel' className='gap-1 py-1 px-2' onClick={ deleteGoal } >
          <DeleteOutlined />
          Remove
        </MBaseButton>
        <MFormButton color='primary' className='gap-1 py-1 px-2' type='submit' role='form'>
          <SaveOutlined />
          Save
        </MFormButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col'>

            <div className='flex flex-col w-full'>
              <div className='flex gap-4 items-center w-full overflow-hidden'>
                <MerchantIcon name={ spending?.name } className='flex-none' />
                <div className='flex flex-col flex-1 overflow-hidden'>
                  <p className='text-ellipsis truncate min-w-0'>
                    { spending?.name }
                  </p>
                  <MSpan weight='semibold'>
                    { locale.formatAmount(spending?.currentAmount, AmountType.Stored) }
                    <span className='font-normal'>of</span>
                    { locale.formatAmount(spending?.targetAmount, AmountType.Stored) }
                  </MSpan>
                </div>
              </div>
              <div className='w-full bg-gray-200 rounded-full h-1.5 my-2 dark:bg-gray-700 relative'>
                <div
                  className='absolute top-0 bg-green-600 h-1.5 rounded-full dark:bg-green-600'
                  style={ { width: `${allocatedProgress}%` } }
                />
                <div
                  className='absolute top-0 bg-blue-600 h-1.5 rounded-full dark:bg-blue-600'
                  style={ { width: `${usedProgress}%` } }
                />
              </div>
            </div>

            <MDivider className='w-1/2 my-4' />

            <MTextField className='w-full' label='Expense' name='name' required />
            <MAmountField allowNegative={ false } className='w-full' label='Amount' name='amount' required />
            <MDatePicker
              label='Target Date'
              name='nextRecurrence'
              className='w-full'
              required
              min={ startOfToday() }
            />
            <MSelectFunding
              className='w-full'
              label='When do you want to fund the expense?'
              menuPortalTarget={ document.body }
              name='fundingScheduleId'
              required
            />
            <MCheckbox
              id='goal-details-paused'
              data-testid='goal-details-paused'
              name='isPaused'
              label='Paused?'
              description='Pause this goal to temporarily stop contributions to it.'
            />
          </div>
          <MDivider className='block md:hidden w-1/2' />
          <div className='w-full md:w-1/2 flex flex-col gap-2'>
            <MSpan className='text-xl my-2'>
              Goal Timeline
            </MSpan>
            <GoalTimeline spendingId={ spending.spendingId } />
          </div>
        </div>
      </div>
    </MForm>
  );
}
