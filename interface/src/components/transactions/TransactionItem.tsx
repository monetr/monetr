/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import MSelectSpendingTransaction from '@monetr/interface/components/MSelectSpendingTransaction';
import MSpan from '@monetr/interface/components/MSpan';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import Transaction from '@monetr/interface/models/Transaction';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface TransactionItemProps {
  transaction: Transaction;
}

export default function TransactionItem({ transaction }: TransactionItemProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: spending } = useSpending(transaction.spendingId);
  const detailsUrl: string = `/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`;

  const amountClassnames = mergeTailwind(
    {
      'dark:text-dark-monetr-green': transaction.getIsAddition(),
      'dark:text-dark-monetr-red': !transaction.getIsAddition(),
    },
    'text-end',
    'font-semibold',
  );

  interface BudgetingInfoProps {
    className: string;
  }

  function BudgetingInfo(props: BudgetingInfoProps): JSX.Element {
    const className = mergeTailwind(
      'overflow-hidden',
      'text-ellipsis',
      'whitespace-nowrap',
      'min-w-0',
      props.className,
    );

    const spentFromClasses = mergeTailwind(
      {
        // Transaction does have spending
        'font-bold': Boolean(transaction.spendingId),
        'dark:text-dark-monetr-content-emphasis': Boolean(transaction.spendingId),
        // No spending for the transaction
        'font-medium': !transaction.spendingId,
        'dark:text-dark-monetr-content': !transaction.spendingId,
      },
      'md:text-base',
      'min-w-0',
      'overflow-hidden',
      'text-ellipsis',
      'text-sm',
      'whitespace-nowrap',
    );

    if (transaction.getIsAddition()) {
      return (
        <span className={ className }>
          <span className='min-w-0 flex-none truncate font-medium dark:text-dark-monetr-content-subtle'>
            Contribution
          </span>
        </span>
      );
    }

    return (
      <span className={ className }>
        <span className='min-w-0 flex-none truncate font-medium dark:text-dark-monetr-content'>
          Spent from
        </span>
        &nbsp;
        <span className={ spentFromClasses }>
          {spending?.name || 'Free-To-Use'}
        </span>
      </span>
    );
  }

  return (
    <li
      className='group relative w-full px-1 md:px-2'
      id={ transaction.transactionId }
      data-testid={ transaction.transactionId }
    >
      <Link
        className='absolute left-0 top-0 flex h-full w-full cursor-pointer md:hidden md:cursor-auto'
        to={ detailsUrl }
      />
      <div className='group flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
        <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
          <TransactionMerchantIcon name={ transaction.getName() } pending={ transaction.isPending } />
          <div className='flex min-w-0 flex-col overflow-hidden'>
            <MSpan size='md' weight='semibold' color='emphasis' ellipsis>
              { transaction.getName() }
            </MSpan>
            <span className='hidden w-full min-w-0 truncate text-sm font-medium dark:text-dark-monetr-content md:block'>
              { transaction.getMainCategory() }
            </span>
            <BudgetingInfo className='flex w-full text-sm md:hidden' />
          </div>
        </div>
        {!transaction.getIsAddition() && (
          <MSelectSpendingTransaction transaction={ transaction } />
        )}
        {transaction.getIsAddition() && (
          <BudgetingInfo className='hidden md:flex w-1/2 flex-1 items-center pl-6' />
        )}
        <div className='flex shrink-0 items-center justify-end gap-2 md:min-w-[8em]'>
          <span className={ amountClassnames }>
            { locale.formatAmount(Math.abs(transaction.amount), AmountType.Stored, transaction.amount < 0) }
          </span>
          <ArrowLink to={ detailsUrl } />
        </div>
      </div>
    </li>
  );
}
