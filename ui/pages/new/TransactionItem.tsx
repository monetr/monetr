/* eslint-disable max-len */
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { KeyboardArrowRight } from '@mui/icons-material';

import TransactionMerchantIcon from './TransactionMerchantIcon';

import { useSpending } from 'hooks/spending';
import Transaction from 'models/Transaction';
import mergeTailwind from 'util/mergeTailwind';

export interface TransactionItemProps {
  transaction: Transaction;
}

export default function TransactionItem({ transaction }: TransactionItemProps): JSX.Element {
  const spending = useSpending(transaction.spendingId);
  const navigate = useNavigate();

  const amountClassnames = mergeTailwind(
    {
      'text-green-500': transaction.getIsAddition(),
      'text-red-500': !transaction.getIsAddition(),
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
      'text-sm',
      'md:text-base',
      'text-ellipsis',
      'whitespace-nowrap',
      'overflow-hidden',
      'min-w-0',
    );

    if (transaction.getIsAddition()) {
      return (
        <span className={ className }>
          <span className='flex-none dark:text-dark-monetr-content-subtle font-medium text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
            Contribution
          </span>
        </span>
      );
    }

    return (
      <span className={ className }>
        <span className='flex-none dark:text-dark-monetr-content font-medium text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
          Spent from
        </span>
        &nbsp;
        <span className={ spentFromClasses }>
          { spending?.name || 'Free-To-Use' }
        </span>
      </span>
    );
  }

  return (
    <li className='w-full px-1 md:px-2'>
      <div className='flex rounded-lg hover:bg-zinc-600 gap-1 md:gap-4 group px-2 py-1 h-full cursor-pointer md:cursor-auto'>
        <div className='w-full md:w-1/2 flex flex-row gap-4 items-center flex-1 min-w-0'>
          <TransactionMerchantIcon name={ transaction.getName() } pending={ transaction.isPending } />
          <div className='flex flex-col overflow-hidden min-w-0'>
            <span className='text-zinc-50 font-semibold text-base w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              {transaction.getName()}
            </span>
            <span className='hidden md:block dark:text-dark-monetr-content font-medium text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { transaction.getMainCategory() }
            </span>
            <BudgetingInfo className='flex md:hidden text-sm w-full' />
          </div>
        </div>
        <BudgetingInfo className='hidden md:flex w-1/2 flex-1 items-center' />
        <div className='flex md:min-w-[8em] shrink-0 justify-end gap-2 items-center'>
          <span className={ amountClassnames }>
            { transaction.getAmountString() }
          </span>
          <KeyboardArrowRight
            className='dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis flex-none md:cursor-pointer'
            onClick={ openDetails }
          />
        </div>
      </div>
    </li>
  );
}
