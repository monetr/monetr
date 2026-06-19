import { useCallback } from 'react';
import { startOfDay } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { HeartCrack, Save, ShoppingCart, Trash } from 'lucide-react';
import { useParams } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
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
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import RemoveTransactionButton from '@monetr/interface/components/transactions/RemoveTransactionButton';
import SimilarTransactions from '@monetr/interface/components/transactions/SimilarTransactions';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchTransaction } from '@monetr/interface/hooks/usePatchTransaction';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransaction } from '@monetr/interface/hooks/useTransaction';
import type { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './details.module.scss';

interface TransactionValues {
  name: string;
  originalName: string;
  date: Date;
  spendingId: ID<Spending> | null;
  isPending: boolean;
  amount: number;
}

export default function TransactionDetails(): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const selectedBankAccountId = useSelectedBankAccountId();
  const { data: link, isLoading: linkIsLoading } = useCurrentLink();
  const { enqueueSnackbar } = useSnackbar();
  const { transactionId: id } = useParams<{ transactionId: string }>();
  const patchTransaction = usePatchTransaction();
  const transactionId = id || undefined;
  const { data: transaction, isLoading, isError } = useTransaction(transactionId);
  const submit = useCallback(
    async (values: TransactionValues, helpers: FormikHelpers<TransactionValues>) => {
      // We cannot build the updated transaction without a locale to convert the friendly amount back into a stored
      // amount, or without the existing transaction to spread the rest of the fields from. The form is only reachable
      // once both have loaded so this should never actually happen.
      if (!locale || !transaction) {
        return;
      }

      helpers.setSubmitting(true);
      return await patchTransaction({
        transactionId: transaction.transactionId,
        bankAccountId: transaction.bankAccountId,
        name: values.name,
        spendingId: values.spendingId,
        ...(link?.getIsManual() && {
          amount: locale.friendlyToAmount(values.amount),
          date: startOfDay(values.date, {
            in: inTimezone,
          }),
          isPending: values.isPending,
        }),
      })
        .then(() =>
          enqueueSnackbar('Updated transaction successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
        )
        .catch((error: ApiError<APIError>) =>
          enqueueSnackbar(error?.response?.data?.error || 'Failed to update transaction', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [enqueueSnackbar, locale, transaction, patchTransaction, inTimezone, link?.getIsManual],
  );

  if (isLoading || linkIsLoading) {
    return (
      <MForm
        className={styles.form}
        enableReinitialize={true}
        initialValues={{} as TransactionValues}
        onSubmit={submit}
      >
        <MTopNavigation
          base={`/bank/${selectedBankAccountId}/transactions`}
          breadcrumb={transaction?.name ?? undefined}
          icon={ShoppingCart}
          title='Transactions'
        />
        <div className={styles.body}>
          <div className={styles.columns}>
            <div className={styles.column}>
              <Flex justify='center'>
                <MerchantIcon name={transaction?.name ?? undefined} />
              </Flex>
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
              <div className={styles.formButtons}>
                <Button disabled variant='destructive'>
                  <Trash />
                  Remove
                </Button>
                <Button disabled variant='primary'>
                  <Save />
                  Save Changes
                </Button>
              </div>
            </div>
          </div>
        </div>
      </MForm>
    );
  }

  if (!transactionId && !isLoading) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>There wasn&apos;t an expense specified...</Typography>
      </div>
    );
  }
  // isLoading is already false by this point (the loading branch above returns early), so this just guards the error
  // case and the absence of the transaction or locale. Narrowing both here lets us safely read them below.
  if (isError || !transaction || !locale) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>We weren&apos;t able to load details for the transaction specified...</Typography>
      </div>
    );
  }

  const initialValues: TransactionValues = {
    name: transaction.name ?? '',
    originalName: transaction.originalName,
    date: transaction.date,
    spendingId: transaction.spendingId ?? null,
    isPending: transaction.isPending,
    amount: locale.amountToFriendly(transaction.amount),
  };

  return (
    <MForm className={styles.form} enableReinitialize={true} initialValues={initialValues} onSubmit={submit}>
      <MTopNavigation
        base={`/bank/${transaction.bankAccountId}/transactions`}
        breadcrumb={transaction?.name ?? undefined}
        icon={ShoppingCart}
        title='Transactions'
      />
      <div className={styles.body}>
        <div className={styles.columns}>
          <div className={styles.column}>
            <Flex justify='center'>
              <MerchantIcon name={transaction?.name ?? undefined} />
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
            <div className={styles.formButtons}>
              <RemoveTransactionButton transaction={transaction} />
              <FormButton role='form' type='submit' variant='primary'>
                <Save />
                Save Changes
              </FormButton>
            </div>
          </div>
          <div className={styles.column}>
            <SimilarTransactions transaction={transaction} />
          </div>
        </div>
      </div>
    </MForm>
  );
}
