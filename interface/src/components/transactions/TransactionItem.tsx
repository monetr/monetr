import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';

import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionItemSelectSpending from '@monetr/interface/components/transactions/TransactionItemSelectSpending';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import type Transaction from '@monetr/interface/models/Transaction';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import itemStyles from './TransactionItem.module.scss';
import selectSpendingStyles from './TransactionItemSelectSpending.module.scss';

export interface TransactionItemProps {
  transaction: Transaction;
}

export default function TransactionItem({ transaction }: TransactionItemProps): JSX.Element {
  const detailsUrl: string = `/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`;

  return (
    <li
      className={mergeTailwind(itemStyles.root, selectSpendingStyles.transactionItemRoot)}
      data-testid={transaction.transactionId}
      id={transaction.transactionId}
    >
      <Link className={itemStyles.mobileLink} to={detailsUrl} />
      <div className={itemStyles.inner}>
        <div className={itemStyles.leftSection}>
          <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
          <Flex flex='shrink' gap='none' orientation='column'>
            <Typography color='emphasis' ellipsis size='md' weight='semibold'>
              {transaction.getName()}
            </Typography>
            <Typography className={itemStyles.categoryLabel} ellipsis size='sm' weight='medium'>
              {transaction.getMainCategory()}
            </Typography>
            <BudgetingInfo className={itemStyles.budgetingMobile} transaction={transaction} />
          </Flex>
        </div>
        {!transaction.getIsAddition() && <TransactionItemSelectSpending transaction={transaction} />}
        {transaction.getIsAddition() && (
          <BudgetingInfo className={itemStyles.budgetingAddition} transaction={transaction} />
        )}
        <div className={itemStyles.amountSection}>
          <TransactionAmount transaction={transaction} />
          <Link className={itemStyles.arrowLink} tabIndex={-1} to={detailsUrl}>
            <ChevronRight />
          </Link>
        </div>
      </div>
    </li>
  );
}

interface BudgetingInfoProps {
  className: string;
  transaction: Transaction;
}

function BudgetingInfo({ transaction, className }: BudgetingInfoProps): JSX.Element {
  const { data: spending } = useSpending(transaction.spendingId);

  if (transaction.getIsAddition()) {
    return (
      <span className={mergeTailwind(itemStyles.budgetingInfo, className)}>
        <span className={itemStyles.contributionLabel}>Contribution</span>
      </span>
    );
  }

  return (
    <span className={mergeTailwind(itemStyles.budgetingInfo, className)}>
      <span className={itemStyles.spentFromLabel}>Spent from</span>
      &nbsp;
      <span className={itemStyles.spentFromValue} data-hasspending={String(Boolean(transaction.spendingId))}>
        {spending?.name || 'Free-To-Use'}
      </span>
    </span>
  );
}
