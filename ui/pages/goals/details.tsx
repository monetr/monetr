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
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MAmountField from 'components/MAmountField';
import MFormButton, { MBaseButton } from 'components/MButton';
import MCheckbox from 'components/MCheckbox';
import MDatePicker from 'components/MDatePicker';
import MForm from 'components/MForm';
import MSelectFunding from 'components/MSelectFunding';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import MTopNavigation from 'components/MTopNavigation';
import { startOfDay, startOfToday } from 'date-fns';
import { useRemoveSpending, useSpending, useUpdateSpending } from 'hooks/spending';
import { showTransferModal } from 'modals/TransferModal';
import Spending, { SpendingType } from 'models/Spending';
import MerchantIcon from 'pages/new/MerchantIcon';
import { amountToFriendly, friendlyToAmount } from 'util/amounts';
import { APIError } from 'util/request';

interface GoalValues {
  name: string;
  amount: number;
  nextRecurrence: Date;
  fundingScheduleId: number;
  isPaused: boolean;
}

export default function GoalDetails(): JSX.Element {
  const removeSpending = useRemoveSpending();
  const updateSpending = useUpdateSpending();
  const navigate = useNavigate();
  const { spendingId } = useParams();
  const { enqueueSnackbar } = useSnackbar();
  const { data: spending, isLoading, isError } = useSpending(spendingId && +spendingId);

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
      recurrenceRule: null,
      targetAmount: friendlyToAmount(values.amount),
      isPaused: values.isPaused,
    });

    return updateSpending(updatedSpending)
      .catch((error: AxiosError<APIError>) => {
        const message = error.response.data.error || 'Failed to update expense.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      })
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: GoalValues = {
    name: spending.name,
    amount: amountToFriendly(spending.targetAmount),
    nextRecurrence: spending.nextRecurrence,
    fundingScheduleId: spending.fundingScheduleId,
    isPaused: spending.isPaused,
  };

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
            <div className='w-full flex justify-center mb-2'>
              <MerchantIcon name={ spending?.name } />
            </div>
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
              name="isPaused"
              label="Paused?"
              description="Pause this goal to temporarily stop contributions to it."
            />
          </div>
        </div>
      </div>
    </MForm>
  );
}
