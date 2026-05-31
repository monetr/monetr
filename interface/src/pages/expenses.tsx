import { Fragment } from 'react';
import { HeartCrack, Plus, Receipt } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import ExpenseItem from '@monetr/interface/components/expenses/ExpenseItem';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useSpendingFiltered } from '@monetr/interface/hooks/useSpendingFiltered';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { SpendingType } from '@monetr/interface/models/Spending';

import styles from './expenses.module.scss';

export default function Expenses(): JSX.Element {
  const { data: expenses, isError, isLoading } = useSpendingFiltered(SpendingType.Expense);

  if (isLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>We weren't able to retrieve expenses at this time...</Typography>
      </div>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={Receipt} title='Expenses'>
        <Button onClick={showNewExpenseModal} variant='primary'>
          <Plus />
          New Expense
        </Button>
      </MTopNavigation>
      <div className={styles.content}>
        {(expenses ?? []).length === 0 && <EmptyState />}
        {(expenses?.length ?? 0) > 0 && (
          <ul className={styles.list}>
            {expenses
              ?.sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
              .map(item => (
                <ExpenseItem key={item.spendingId} spending={item} />
              ))}
          </ul>
        )}
      </div>
    </Fragment>
  );
}

function EmptyState(): JSX.Element {
  return (
    <div className={styles.empty}>
      <div className={styles.emptyInner}>
        <div className={styles.iconRow}>
          <Receipt className={styles.emptyIcon} />
        </div>
        <Typography align='center' color='subtle' size='xl'>
          You don't have any expenses yet...
        </Typography>
        <Typography align='center' color='subtle' size='lg'>
          Expenses are budgets for recurring spending. Things like your streaming subscription, rent, or car payments.
        </Typography>
      </div>
    </div>
  );
}
