import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { DeleteOutlined, HeartBroken, PriceCheckOutlined, SaveOutlined, SwapVertOutlined } from '@mui/icons-material';
import { AxiosError } from 'axios';
import { startOfDay, startOfToday } from 'date-fns';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import ExpenseTimeline from './ExpenseTimeline';
import FormButton from '@monetr/interface/components/FormButton';
import MAmountField from '@monetr/interface/components/MAmountField';
import { MBaseButton } from '@monetr/interface/components/MButton';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MDivider from '@monetr/interface/components/MDivider';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
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

interface ExpenseValues {
  name: string;
  amount: number;
  nextRecurrence: Date;
  fundingScheduleId: string;
  recurrenceRule: string;
}

export default function ExpenseDetails(): JSX.Element {
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
          There wasn't an expense specified...
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
          Couldn't find the expense you specified...
        </MSpan>
      </div>
    );
  }

  if (!spending) {
    return null;
  }

  if (spending.spendingType !== SpendingType.Expense) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          This spending object is not an expense...
        </MSpan>
      </div>
    );
  }

  function backToExpenses() {
    navigate(`/bank/${spending.bankAccountId}/expenses`);
  }

  async function deleteExpense(): Promise<void> {
    if (!spending) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete expense: ${ spending.name }`)) {
      return removeSpending(spending.spendingId)
        .then(() => backToExpenses());
    }

    return Promise.resolve();
  }

  async function submit(values: ExpenseValues, helpers: FormikHelpers<ExpenseValues>): Promise<void> {
    helpers.setSubmitting(true);

    const updatedSpending = new Spending({
      ...spending,
      name: values.name,
      description: null,
      nextRecurrence: startOfDay(values.nextRecurrence),
      fundingScheduleId: values.fundingScheduleId,
      ruleset: values.recurrenceRule,
      targetAmount: locale.friendlyToAmount(values.amount),
    });

    return updateSpending(updatedSpending)
      .then(() => void enqueueSnackbar(
        'Updated expense successfully',
        {
          variant: 'success',
          disableWindowBlurListener: true,
        },
      ))
      .catch((error: AxiosError<APIError>) => void enqueueSnackbar(
        error.response.data.error || 'Failed to update expense',
        {
          variant: 'error',
          disableWindowBlurListener: true,
        },
      ))
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: ExpenseValues = {
    name: spending.name,
    amount: locale.amountToFriendly(spending.targetAmount),
    nextRecurrence: spending.nextRecurrence,
    fundingScheduleId: spending.fundingScheduleId,
    recurrenceRule: spending.ruleset,
  };

  const progress = ((Math.min(
    spending?.currentAmount,
    spending?.targetAmount,
  ) / spending?.targetAmount) * 100).toFixed(0);

  return (
    <MForm initialValues={ initialValues } onSubmit={ submit } className='flex w-full h-full flex-col'>
      <MTopNavigation
        icon={ PriceCheckOutlined }
        title='Expenses'
        base={ `/bank/${spending.bankAccountId}/expenses` }
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
        <MBaseButton color='cancel' className='gap-1 py-1 px-2' onClick={ deleteExpense } >
          <DeleteOutlined />
          Remove
        </MBaseButton>
        <FormButton variant='primary' type='submit' role='form'>
          <SaveOutlined />
          Save
        </FormButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>

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
                  style={ { width: `${progress}%` } }
                />
              </div>
            </div>

            <MDivider className='w-1/2 my-4' />

            <MTextField className='w-full' label='Expense' name='name' required data-1p-ignore />
            <MAmountField allowNegative={ false } className='w-full' label='Amount' name='amount' required />
            <MDatePicker
              label='Next Occurrence'
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
            <MSelectFrequency
              className='w-full'
              dateFrom='nextRecurrence'
              label='How often do you need this expense?'
              name='recurrenceRule'
              placeholder='Select a spending frequency...'
              required
            />
          </div>
          <MDivider className='block md:hidden w-1/2' />
          <div className='w-full md:w-1/2 flex flex-col gap-2'>
            <MSpan className='text-xl my-2'>
              Expense Timeline
            </MSpan>
            <ExpenseTimeline spendingId={ spending.spendingId } />
          </div>
        </div>
      </div>
    </MForm>
  );
}
