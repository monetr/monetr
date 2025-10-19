import { Fragment, useCallback, useEffect, useRef } from 'react';
import { useNavigationType } from 'react-router-dom';
import { HeartBroken } from '@mui/icons-material';
import { Plus, Receipt } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import ExpenseItem from '@monetr/interface/components/expenses/ExpenseItem';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useSpendingFiltered } from '@monetr/interface/hooks/useSpendingFiltered';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { SpendingType } from '@monetr/interface/models/Spending';

let evilScrollPosition: number = 0;

export default function Expenses(): JSX.Element {
  const { data: expenses, isError, isLoading } = useSpendingFiltered(SpendingType.Expense);

  // Scroll restoration code.
  const ref = useRef<HTMLDivElement>(null);
  const navigationType = useNavigationType();
  const onScroll = useCallback(() => {
    evilScrollPosition = ref.current.scrollTop;
  }, [ref]);
  useEffect(() => {
    if (!ref.current) {
      return undefined;
    }

    if (navigationType === 'POP') {
      ref.current.scrollTop = evilScrollPosition;
    }
    const current = ref.current;
    ref.current.addEventListener('scroll', onScroll);
    return () => {
      current.removeEventListener('scroll', onScroll);
    };
    // Fix bug with current impl.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ref.current, navigationType, onScroll]);

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>One moment...</MSpan>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>Something isn't right...</MSpan>
        <MSpan className='text-2xl'>We weren't able to retrieve expenses at this time...</MSpan>
      </div>
    );
  }

  function ListContent(): JSX.Element {
    if (expenses.length === 0) {
      return <EmptyState />;
    }

    return (
      <ul className='w-full flex flex-col gap-2 py-2 pb-16'>
        {expenses
          ?.sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
          .map(item => (
            <ExpenseItem spending={item} key={item.spendingId} />
          ))}
      </ul>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={Receipt} title='Expenses'>
        <Button variant='primary' className='gap-1 py-1 px-2' onClick={showNewExpenseModal}>
          <Plus />
          New Expense
        </Button>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0' ref={ref}>
        <ListContent />
      </div>
    </Fragment>
  );
}

function EmptyState(): JSX.Element {
  return (
    <div className='w-full h-full flex justify-center items-center'>
      <div className='flex flex-col gap-2 items-center max-w-md'>
        <div className='w-full flex justify-center space-x-4'>
          <Receipt className='dark:text-dark-monetr-content-muted h-12 w-12' />
        </div>
        <MSpan size='xl' color='subtle' className='text-center'>
          You don't have any expenses yet...
        </MSpan>
        <MSpan size='lg' color='subtle' className='text-center'>
          Expenses are budgets for recurring spending. Things like your streaming subscription, rent, or car payments.
        </MSpan>
      </div>
    </div>
  );
}
