/* eslint-disable max-len */
import React from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { ArrowBackOutlined, HeartBroken, PriceCheckOutlined, SaveOutlined } from '@mui/icons-material';

import ExpenseTimeline from './ExpenseTimeline';

import { MBaseButton } from 'components/MButton';
import MDivider from 'components/MDivider';
import MForm from 'components/MForm';
import MSelect from 'components/MSelect';
import MSelectFunding from 'components/MSelectFunding';
import MSidebarToggle from 'components/MSidebarToggle';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useSpending } from 'hooks/spending';
import MerchantIcon from 'pages/new/MerchantIcon';

interface ExpenseValues {
  name: string;
  amount: number;
  nextRecurrence: moment.Moment;
  fundingScheduleId: number;
  recurrenceRule: string;
}

export default function ExpenseDetails(): JSX.Element {
  const navigate = useNavigate();
  const { spendingId } = useParams();

  const spending = useSpending(spendingId && +spendingId);

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

  if (!spending) {
    return null;
  }

  function submit() {

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
          <MBaseButton
            color='cancel'
            className='gap-1 py-1 px-2'
            onClick={ () => navigate(`/bank/${spending.bankAccountId}/expenses`) }
          >
            <ArrowBackOutlined />
            Cancel
          </MBaseButton>
          <MBaseButton color='primary' className='gap-1 py-1 px-2'>
            <SaveOutlined />
          Save Changes
          </MBaseButton>
        </div>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <div className='w-full flex justify-center mb-2'>
              <MerchantIcon name={ spending?.name } />
            </div>
            <MTextField
              label='Expense'
              name='name'
              value={ spending?.name }
              className='w-full'
            />
            <MTextField
              label='Amount'
              name='amount'
              prefix='$'
              type='number'
              className='w-full'
            />
            <MTextField
              label='Next Occurrence'
              name='nextRecurrence'
              type='date'
              className='w-full'
            />
            <MSelectFunding
              menuPortalTarget={ document.body }
              label='When do you want to fund the expense?'
              name='fundingScheduleId'
              className='w-full'
            />
            <MSelect
              label='Spending Frequency'
              name='recurrenceRule'
              placeholder='Select a spending frequency...'
              options={ [] }
              className='w-full'
            />
            <MDivider className='w-1/2' />
            <MSpan className='text-xl my-2'>
            Stats
            </MSpan>
            <div className='w-full'>
              <MSpan>Estimated Regular Contribution Amount:</MSpan>
              &nbsp;
              <MSpan>$140.00</MSpan>
            </div>
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
