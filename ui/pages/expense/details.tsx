/* eslint-disable max-len */
import React from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { ArrowBackOutlined, DeleteOutlined, HeartBroken, PriceCheckOutlined, SaveOutlined, SwapVertOutlined } from '@mui/icons-material';
import { AxiosError } from 'axios';
import { FormikHelpers } from 'formik';
import moment from 'moment';
import { useSnackbar } from 'notistack';

import ExpenseTimeline from './ExpenseTimeline';

import MAmountField from 'components/MAmountField';
import MFormButton, { MBaseButton } from 'components/MButton';
import MDivider from 'components/MDivider';
import MForm from 'components/MForm';
import MSelectFrequency from 'components/MSelectFrequency';
import MSelectFunding from 'components/MSelectFunding';
import MSidebarToggle from 'components/MSidebarToggle';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useRemoveSpending, useSpending, useUpdateSpending } from 'hooks/spending';
import { showTransferModal } from 'modals/TransferModal';
import Spending, { SpendingType } from 'models/Spending';
import MerchantIcon from 'pages/new/MerchantIcon';
import { APIError } from 'util/request';

interface ExpenseValues {
  name: string;
  amount: number;
  nextRecurrence: moment.Moment;
  fundingScheduleId: number;
  recurrenceRule: string;
}

export default function ExpenseDetails(): JSX.Element {
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
      nextRecurrence: moment(values.nextRecurrence).startOf('day'),
      fundingScheduleId: values.fundingScheduleId,
      recurrenceRule: values.recurrenceRule,
      targetAmount: Math.ceil(values.amount * 100),
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

  const initialValues: ExpenseValues = {
    name: spending.name,
    amount: +(spending.targetAmount / 100).toFixed(2),
    nextRecurrence: spending.nextRecurrence,
    fundingScheduleId: spending.fundingScheduleId,
    recurrenceRule: spending.recurrenceRule,
  };

  return (
    <MForm initialValues={ initialValues } onSubmit={ submit } className='flex w-full h-full flex-col'>
      <div className='w-full h-auto md:h-12 flex flex-col md:flex-row md:items-center px-4 gap-4 md:justify-between'>
        <div className='flex items-center gap-2 mt-2 md:mt-0'>
          <MSidebarToggle />
          <span className='flex items-center text-2xl dark:text-dark-monetr-content-subtle font-bold'>
            <PriceCheckOutlined />
          </span>
          <Link
            className='text-2xl hidden md:block dark:text-dark-monetr-content-subtle dark:hover:text-dark-monetr-content-emphasis font-bold cursor-pointer'
            to={ `/bank/${spending.bankAccountId}/expenses` }
          >
            Expenses
          </Link>
          <span className='text-2xl hidden md:block dark:text-dark-monetr-content-subtle font-bold'>
          /
          </span>
          <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold'>
            { spending?.name }
          </span>
        </div>
        <div className='flex gap-2'>
          <MBaseButton color='secondary' className='gap-1 py-1 px-2' onClick={ backToExpenses } >
            <ArrowBackOutlined />
            Cancel
          </MBaseButton>
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
          <MFormButton color='primary' className='gap-1 py-1 px-2' type='submit' role='form'>
            <SaveOutlined />
            Save Changes
          </MFormButton>
        </div>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <div className='w-full flex justify-center mb-2'>
              <MerchantIcon name={ spending?.name } />
            </div>
            <MTextField className='w-full' label='Expense' name='name' />
            <MAmountField allowNegative={ false } className='w-full' label='Amount' name='amount' />
            <MTextField
              className='w-full'
              label='Next Occurrence'
              name='nextRecurrence'
              type='date'
            />
            <MSelectFunding
              className='w-full'
              label='When do you want to fund the expense?'
              menuPortalTarget={ document.body }
              name='fundingScheduleId'
            />
            <MSelectFrequency
              className='w-full'
              dateFrom='nextRecurrence'
              label='How often do you need this expense?'
              name='recurrenceRule'
              placeholder='Select a spending frequency...'
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
