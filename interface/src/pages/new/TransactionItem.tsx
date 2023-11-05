/* eslint-disable max-len */
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { KeyboardArrowRight } from '@mui/icons-material';

import TransactionMerchantIcon from './TransactionMerchantIcon';
import MSelectSpendingTransaction from '@monetr/interface/components/MSelectSpendingTransaction';
import { useSpendingOld } from '@monetr/interface/hooks/spending';
import Transaction from '@monetr/interface/models/Transaction';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface TransactionItemProps {
  transaction: Transaction;
}

export default function TransactionItem({ transaction }: TransactionItemProps): JSX.Element {
  const spending = useSpendingOld(transaction.spendingId);
  const navigate = useNavigate();

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

  function openDetails() {
    navigate(`/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`);
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
        'font-bold': !!transaction.spendingId,
        'dark:text-dark-monetr-content-emphasis': !!transaction.spendingId,
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
    <li className='group relative w-full px-1 md:px-2'>
      <div
        className='absolute left-0 top-0 flex h-full w-full cursor-pointer md:hidden md:cursor-auto'
        onClick={ openDetails }
      />
      <div className='group flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
        <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
          <TransactionMerchantIcon name={ transaction.getName() } pending={ transaction.isPending } />
          <div className='flex min-w-0 flex-col overflow-hidden'>
            <span className='w-full min-w-0 truncate text-base font-semibold dark:text-dark-monetr-content-emphasis'>
              {transaction.getName()}
            </span>
            <span className='hidden w-full min-w-0 truncate text-sm font-medium dark:text-dark-monetr-content md:block'>
              {transaction.getMainCategory()}
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
            {transaction.getAmountString()}
          </span>
          <KeyboardArrowRight
            className='flex-none dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis md:cursor-pointer'
            onClick={ openDetails }
          />
        </div>
      </div>
    </li>
  );
}
