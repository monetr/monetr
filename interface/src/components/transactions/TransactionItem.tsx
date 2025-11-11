import { Link } from 'react-router-dom';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import Flex from '@monetr/interface/components/Flex';
import MSelectSpendingTransaction from '@monetr/interface/components/MSelectSpendingTransaction';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import type Transaction from '@monetr/interface/models/Transaction';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface TransactionItemProps {
  transaction: Transaction;
}

export default function TransactionItem({ transaction }: TransactionItemProps): JSX.Element {
  const detailsUrl: string = `/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`;

  return (
    <li
      className='group relative w-full px-1 md:px-2'
      id={transaction.transactionId}
      data-testid={transaction.transactionId}
    >
      <Link
        className='absolute left-0 top-0 flex h-full w-full cursor-pointer md:hidden md:cursor-auto'
        to={detailsUrl}
      />
      <div className='group flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
        <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
          <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
          <Flex orientation='column' gap='none' flex='shrink'>
            <Typography size='md' weight='semibold' color='emphasis' ellipsis>
              {transaction.getName()}
            </Typography>
            <Typography size='sm' weight='medium' ellipsis className='hidden md:block'>
              {transaction.getMainCategory()}
            </Typography>
            <BudgetingInfo className='flex w-full text-sm md:hidden' transaction={transaction} />
          </Flex>
        </div>
        {!transaction.getIsAddition() && <MSelectSpendingTransaction transaction={transaction} />}
        {transaction.getIsAddition() && (
          <BudgetingInfo className='hidden md:flex w-1/2 flex-1 items-center pl-6' transaction={transaction} />
        )}
        <div className='flex shrink-0 items-center justify-end gap-2 md:min-w-[8em]'>
          <TransactionAmount transaction={transaction} />
          <ArrowLink to={detailsUrl} />
        </div>
      </div>
    </li>
  );
}

interface BudgetingInfoProps {
  className: string;
  transaction: Transaction;
}

function BudgetingInfo({ transaction, ...props }: BudgetingInfoProps): JSX.Element {
  const { data: spending } = useSpending(transaction.spendingId);
  const className = mergeTailwind('overflow-hidden', 'text-ellipsis', 'whitespace-nowrap', 'min-w-0', props.className);

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
      <span className={className}>
        <span className='min-w-0 flex-none truncate font-medium dark:text-dark-monetr-content-subtle'>
          Contribution
        </span>
      </span>
    );
  }

  return (
    <span className={className}>
      <span className='min-w-0 flex-none truncate font-medium dark:text-content'>Spent from</span>
      &nbsp;
      <span className={spentFromClasses}>{spending?.name || 'Free-To-Use'}</span>
    </span>
  );
}
