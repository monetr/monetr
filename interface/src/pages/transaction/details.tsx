import React from 'react';
import { useParams } from 'react-router-dom';
import { ShoppingCartOutlined } from '@mui/icons-material';
import { AxiosError } from 'axios';
import { startOfDay } from 'date-fns';
import { FormikHelpers } from 'formik';
import { HeartCrack, Save } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import MAmountField from '@monetr/interface/components/MAmountField';
import MCheckbox from '@monetr/interface/components/MCheckbox';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import SimilarTransactions from '@monetr/interface/components/transactions/SimilarTransactions';
import { useCurrentLink } from '@monetr/interface/hooks/links';
import { useTransaction, useUpdateTransaction } from '@monetr/interface/hooks/transactions';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import Transaction from '@monetr/interface/models/Transaction';
import { APIError } from '@monetr/interface/util/request';

interface TransactionValues {
  name: string;
  originalName: string;
  date: Date;
  spendingId: string | null;
  isPending: boolean;
  amount: number;
}

export default function TransactionDetails(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: link } = useCurrentLink();
  const { enqueueSnackbar } = useSnackbar();
  const { transactionId: id } = useParams();
  const updateTransaction = useUpdateTransaction();
  const transactionId = id || null;

  const { data: transaction, isLoading, isError } = useTransaction(transactionId);

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  if (!transactionId && !isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          There wasn't an expense specified...
        </MSpan>
      </div>
    );
  }
  if ((isError || !transaction) && !isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          We weren't able to load details for the transaction specified...
        </MSpan>
      </div>
    );
  }

  async function submit(values: TransactionValues, helpers: FormikHelpers<TransactionValues>) {
    const updatedTransaction = new Transaction({
      ...transaction,
      name: values.name,
      spendingId: values.spendingId,
      amount: locale.friendlyToAmount(values.amount),
      date: startOfDay(values.date),
      isPending: values.isPending,
    });

    helpers.setSubmitting(true);
    return updateTransaction(updatedTransaction)
      .then(() => enqueueSnackbar(
        'Updated transaction successfully',
        {
          variant: 'success',
          disableWindowBlurListener: true,
        },
      ))
      .catch((error: AxiosError<APIError>) => enqueueSnackbar(
        error?.response?.data?.error || 'Failed to update transaction',
        {
          variant: 'error',
          disableWindowBlurListener: true,
        },
      ))
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: TransactionValues = {
    name: transaction.name,
    originalName: transaction.originalName,
    date: transaction.date,
    spendingId: transaction.spendingId,
    isPending: transaction.isPending,
    amount: locale.amountToFriendly(transaction.amount),
  };

  return (
    <MForm
      initialValues={ initialValues }
      enableReinitialize={ true }
      onSubmit={ submit }
      className='flex w-full h-full flex-col'
    >
      <MTopNavigation
        icon={ ShoppingCartOutlined }
        title='Transactions'
        base={ `/bank/${transaction.bankAccountId}/transactions` }
        breadcrumb={ transaction?.name }
      >
        <Button variant='primary' className='gap-1 py-1 px-2' type='submit'>
          <Save />
          Save Changes
        </Button>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <div className='w-full flex justify-center mb-2'>
              <MerchantIcon name={ transaction?.name } />
            </div>
            <MTextField
              id='transaction-name-search'
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
            <MAmountField
              className='w-full'
              disabled
              label='Amount'
              name='amount'
            />
            <MDatePicker
              label='Date'
              name='date'
              className='w-full'
              disabled={ !link.getIsManual() }
            />
            <MCheckbox
              id='transaction-details-pending'
              data-testid='transaction-details-pending'
              name='isPending'
              label='Is Pending'
              description='Transaction has not yet cleared, the name or amount may change.'
              className='w-full'
              disabled={ !link.getIsManual() }
            />
            { !transaction.getIsAddition() && (
              <MSelectSpending
                className='w-full'
                name='spendingId'
              />
            ) }
          </div>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <SimilarTransactions transaction={ transaction } />
          </div>
        </div>
      </div>
    </MForm>
  );
}
