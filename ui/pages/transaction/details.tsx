/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowBackOutlined, HeartBroken, SaveOutlined, ShoppingCartOutlined } from '@mui/icons-material';

import { MBaseButton } from 'components/MButton';
import MForm from 'components/MForm';
import MSelect from 'components/MSelect';
import MSidebarToggle from 'components/MSidebarToggle';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useTransaction } from 'hooks/transactions';
import MerchantIcon from 'pages/new/MerchantIcon';
import moment from 'moment';
import { Formik } from 'formik';
import MSelectSpending from 'components/MSelectSpending';

interface TransactionValues {
  name: string;
  originalName: string;
  date: moment.Moment;
  spendingId: number | null;
  amount: number;
}

export default function TransactionDetails(): JSX.Element {
  const { transactionId: id } = useParams();
  const transactionId = +id || null;

  const { result: transaction, isLoading, isError } = useTransaction(transactionId);
  if (!transactionId) {
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
  if (isError || !transaction) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          We weren't able to load details for the transaction specified...
        </MSpan>
      </div>
    );
  }

  function submit() {

  }

  const initialValues: TransactionValues = {
    name: transaction.name,
    originalName: transaction.originalName,
    date: transaction.date,
    spendingId: transaction.spendingId,
    amount: +(transaction.amount / 100).toFixed(2),
  };

  return (
    <Formik
      initialValues={ initialValues }
      onSubmit={ submit }
    >
      <MForm className='flex w-full h-full flex-col'>
        <div className='w-full h-auto md:h-12 flex flex-col md:flex-row md:items-center px-4 gap-4 md:justify-between'>
          <div className='flex grow items-center gap-2 mt-2 md:mt-0 min-w-0 overflow-none'>
            <MSidebarToggle />
            <span className='flex items-center text-2xl dark:text-dark-monetr-content-subtle font-bold'>
              <ShoppingCartOutlined />
            </span>
            <Link
              className='text-2xl hidden md:block dark:text-dark-monetr-content-subtle dark:hover:text-dark-monetr-content-emphasis font-bold cursor-pointer'
              to={ `/bank/${transaction?.bankAccountId}/transactions` }
            >
              Transactions
            </Link>
            <span className='text-2xl hidden md:block dark:text-dark-monetr-content-subtle font-bold'>
            /
            </span>
            <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold whitespace-nowrap text-ellipsis overflow-hidden min-w-0 flex-1'>
              { transaction?.name }
            </span>
          </div>
          <div className='md:min-w-0 fixed md:static bottom-2 right-2 h-10 md:h-16 items-center flex gap-2 justify-end'>
            <MBaseButton color='cancel' className='gap-1 py-1 px-2'>
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
                <MerchantIcon name={ transaction?.name } />
              </div>
              <MTextField
                label='Name'
                placeholder='Transaction name...'
                name='name'
                className='w-full'
              />
              <MTextField
                label='Original Name'
                placeholder='No original name...?'
                name='originalName'
                className='w-full'
                disabled
              />
              <MTextField
                label='Amount'
                name='amount'
                prefix='$'
                type='number'
                className='w-full'
                disabled
              />
              <MTextField
                label='Date'
                name='date'
                type='date'
                className='w-full'
                disabled
              />
              <MSelectSpending className='w-full' />
            </div>
          </div>
        </div>
      </MForm>
    </Formik>
  );
}
