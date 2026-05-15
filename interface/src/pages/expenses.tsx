import { Fragment } from 'react';
import { HeartCrack, Plus, Receipt } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import ExpenseItem from '@monetr/interface/components/expenses/ExpenseItem';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useSpendingFiltered } from '@monetr/interface/hooks/useSpendingFiltered';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { SpendingType } from '@monetr/interface/models/Spending';

export default function Expenses(): JSX.Element {
  const { data: expenses, isError, isLoading } = useSpendingFiltered(SpendingType.Expense);

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content size-24' />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>We weren't able to retrieve expenses at this time...</Typography>
      </div>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={Receipt} title='Expenses'>
        <Button className='gap-1 py-1 px-2' onClick={showNewExpenseModal} variant='primary'>
          <Plus />
          New Expense
        </Button>
      </MTopNavigation>
      <div className='w-full flex grow flex-col min-w-0'>
        {(expenses ?? []).length === 0 && <EmptyState />}
        {expenses?.length > 0 && (
          <ul className='w-full flex flex-col gap-2 py-2 pb-16'>
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
    <div className='w-full flex justify-center items-center grow'>
      <div className='flex flex-col gap-2 items-center max-w-md p-2'>
        <div className='w-full flex justify-center space-x-4'>
          <Receipt className='dark:text-dark-monetr-content-muted h-12 w-12' />
        </div>
        <Typography className='text-center' color='subtle' size='xl'>
          You don't have any expenses yet...
        </Typography>
        <Typography className='text-center' color='subtle' size='lg'>
          Expenses are budgets for recurring spending. Things like your streaming subscription, rent, or car payments.
        </Typography>
      </div>
    </div>
  );
}
