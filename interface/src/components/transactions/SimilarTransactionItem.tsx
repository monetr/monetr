
import { format, isThisYear } from 'date-fns';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useTransaction } from '@monetr/interface/hooks/useTransaction';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface SimilarTransactionItemProps {
  transactionId: string;
  /**
   * disableNavigate will remove the arrow link or the click-ability of the similar transaction item.
   */
  disableNavigate?: boolean;
}

export default function SimilarTransactionItem(props: SimilarTransactionItemProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: transaction, isLoading, isError } = useTransaction(props.transactionId);

  if (isLoading) {
    return (
      <li className='group relative w-full px-1 md:px-2'>
        <div className='group animate-pulse flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
          <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
            <div
              className='h-10 w-10 rounded-full dark:bg-dark-monetr-background-subtle'
              aria-label='Transaction Avatar'
            />
            <div className='flex min-w-0 grow flex-col overflow-hidden'>
              <div
                className='w-full rounded-xl h-4 my-1 dark:bg-dark-monetr-background-subtle'
                aria-label='Transaction Name'
              />
              <div
                className='w-1/2 rounded-xl h-3 my-1 dark:bg-dark-monetr-background-subtle opacity-70'
                aria-label='Transaction Date'
              />
            </div>
          </div>
          <div className='flex shrink-0 items-center justify-end gap-2 md:min-w-[8em]'>
            <div
              className='w-1/3 rounded-xl h-4 dark:bg-dark-monetr-background-subtle'
              aria-label='Transaction Amount'
            />
          </div>
        </div>
      </li>
    );
  }

  if (isError) {
    return null;
  }

  const amountClassnames = mergeTailwind(
    {
      'dark:text-dark-monetr-green': transaction.getIsAddition(),
      'dark:text-dark-monetr-red': !transaction.getIsAddition(),
    },
    'text-end',
    'font-semibold',
  );

  const redirectUrl: string = `/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`;

  const dateString = isThisYear(transaction.date)
    ? format(transaction.date, 'MMMM do')
    : format(transaction.date, 'MMMM do, yyyy');

  return (
    <li className='group relative w-full px-1 md:px-2'>
      <div className='group flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
        <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
          <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
          <div className='flex min-w-0 flex-col overflow-hidden'>
            <span className='w-full min-w-0 truncate text-base font-semibold dark:text-dark-monetr-content-emphasis'>
              {transaction.getName()}
            </span>
            <span className='w-full min-w-0 truncate text-sm font-medium dark:text-dark-monetr-content'>
              {dateString}
            </span>
          </div>
        </div>
        <div className='flex shrink-0 items-center justify-end gap-2 md:min-w-[8em]'>
          <span className={amountClassnames}>
            {locale.formatAmount(Math.abs(transaction.amount), AmountType.Stored, transaction.amount < 0)}
          </span>
          {!props.disableNavigate && <ArrowLink to={redirectUrl} />}
        </div>
      </div>
    </li>
  );
}
