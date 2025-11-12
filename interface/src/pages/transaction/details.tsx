import { useCallback } from 'react';
import type { AxiosError } from 'axios';
import { startOfDay } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { HeartCrack, Save, ShoppingCart } from 'lucide-react';
import { useSnackbar } from 'notistack';
import { useParams } from 'react-router-dom';

import Flex from '@monetr/interface/components/Flex';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormCheckbox from '@monetr/interface/components/FormCheckbox';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MForm from '@monetr/interface/components/MForm';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import RemoveTransactionButton from '@monetr/interface/components/transactions/RemoveTransactionButton';
import SimilarTransactions from '@monetr/interface/components/transactions/SimilarTransactions';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransaction } from '@monetr/interface/hooks/useTransaction';
import { useUpdateTransaction } from '@monetr/interface/hooks/useUpdateTransaction';
import Transaction from '@monetr/interface/models/Transaction';
import type { APIError } from '@monetr/interface/util/request';

interface TransactionValues {
  name: string;
  originalName: string;
  date: Date;
  spendingId: string | null;
  isPending: boolean;
  amount: number;
}

export default function TransactionDetails(): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const selectedBankAccountId = useSelectedBankAccountId();
  const { data: link, isLoading: linkIsLoading } = useCurrentLink();
  const { enqueueSnackbar } = useSnackbar();
  const { transactionId: id } = useParams();
  const updateTransaction = useUpdateTransaction();
  const transactionId = id || null;
  const { data: transaction, isLoading, isError } = useTransaction(transactionId);
  const submit = useCallback(
    async (values: TransactionValues, helpers: FormikHelpers<TransactionValues>) => {
      const updatedTransaction = new Transaction({
        ...transaction,
        name: values.name,
        spendingId: values.spendingId,
        amount: locale.friendlyToAmount(values.amount),
        date: startOfDay(values.date, {
          in: inTimezone,
        }),
        isPending: values.isPending,
      });

      helpers.setSubmitting(true);
      return await updateTransaction(updatedTransaction)
        .then(() =>
          enqueueSnackbar('Updated transaction successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
        )
        .catch((error: AxiosError<APIError>) =>
          enqueueSnackbar(error?.response?.data?.error || 'Failed to update transaction', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [enqueueSnackbar, locale, transaction, updateTransaction, inTimezone],
  );

  if (isLoading || linkIsLoading) {
    return (
      <MForm className='flex w-full h-full flex-col' enableReinitialize={true} initialValues={{}} onSubmit={submit}>
        <MTopNavigation
          base={`/bank/${selectedBankAccountId}/transactions`}
          breadcrumb={transaction?.name}
          icon={ShoppingCart}
          title='Transactions'
        />
        <div className='w-full h-full overflow-y-auto min-w-0 p-4 pb-16 md:pb-4'>
          <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
            <div className='w-full md:w-1/2 flex flex-col items-center'>
              <div className='w-full flex justify-center mb-2'>
                <MerchantIcon name={transaction?.name} />
              </div>
              <FormTextField
                autoComplete='off'
                className={layoutVariants({ width: 'full' })}
                data-1p-ignore
                isLoading
                label='Name'
                name='name'
                placeholder='Transaction name...'
              />
              <FormTextField
                autoComplete='off'
                className={layoutVariants({ width: 'full' })}
                disabled
                isLoading
                label='Original Name'
                name='originalName'
                placeholder='No original name...?'
              />
              <FormAmountField
                className={layoutVariants({ width: 'full' })}
                disabled
                isLoading
                label='Amount'
                name='amount'
              />
              <FormDatePicker className={layoutVariants({ width: 'full' })} disabled label='Date' name='date' />
              <FormCheckbox
                className={layoutVariants({ width: 'full' })}
                data-testid='transaction-details-pending'
                description='Transaction has not yet cleared, the name or amount may change.'
                disabled
                label='Is Pending'
                name='isPending'
              />
              <MSelectSpending className={layoutVariants({ width: 'full' })} isLoading name='spendingId' />
            </div>
          </div>
        </div>
      </MForm>
    );
  }

  if (!transactionId && !isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>Something isn't right...</MSpan>
        <MSpan className='text-2xl'>There wasn't an expense specified...</MSpan>
      </div>
    );
  }
  if ((isError || !transaction) && !isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>Something isn't right...</MSpan>
        <MSpan className='text-2xl'>We weren't able to load details for the transaction specified...</MSpan>
      </div>
    );
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
      className='flex w-full h-full flex-col'
      enableReinitialize={true}
      initialValues={initialValues}
      onSubmit={submit}
    >
      <MTopNavigation
        base={`/bank/${transaction.bankAccountId}/transactions`}
        breadcrumb={transaction?.name}
        icon={ShoppingCart}
        title='Transactions'
      >
        <RemoveTransactionButton transaction={transaction} />
        <FormButton role='form' type='submit' variant='primary'>
          <Save />
          Save Changes
        </FormButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4 pb-16 md:pb-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <Flex justify='center'>
              <MerchantIcon name={transaction?.name} />
            </Flex>
            <FormTextField
              autoComplete='off'
              className={layoutVariants({ width: 'full' })}
              data-1p-ignore
              label='Name'
              name='name'
              placeholder='Transaction name...'
            />
            <FormTextField
              autoComplete='off'
              className={layoutVariants({ width: 'full' })}
              disabled
              label='Original Name'
              name='originalName'
              placeholder='No original name...?'
            />
            <FormAmountField className={layoutVariants({ width: 'full' })} disabled label='Amount' name='amount' />
            <FormDatePicker
              className={layoutVariants({ width: 'full' })}
              disabled={!link?.getIsManual()}
              label='Date'
              name='date'
            />
            <FormCheckbox
              className={layoutVariants({ width: 'full' })}
              data-testid='transaction-details-pending'
              description='Transaction has not yet cleared, the name or amount may change.'
              disabled={!link?.getIsManual()}
              label='Is Pending'
              name='isPending'
            />
            {!transaction.getIsAddition() && (
              <MSelectSpending className={layoutVariants({ width: 'full' })} name='spendingId' />
            )}
          </div>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <SimilarTransactions transaction={transaction} />
          </div>
        </div>
      </div>
    </MForm>
  );
}
