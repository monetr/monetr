import React, { Fragment, useCallback, useEffect, useRef } from 'react';
import { useNavigationType } from 'react-router-dom';
import { AddOutlined, HeartBroken, SavingsOutlined } from '@mui/icons-material';

import { MBaseButton } from '@monetr/interface/components/MButton';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useSpendingFiltered } from '@monetr/interface/hooks/spending';
import { showNewGoalModal } from '@monetr/interface/modals/NewGoalModal';
import { SpendingType } from '@monetr/interface/models/Spending';

import GoalItem from './new/GoalItem';

let evilScrollPosition: number = 0;

export default function Goals(): JSX.Element {
  const {
    result: goals,
    isError,
    isLoading,
  } = useSpendingFiltered(SpendingType.Goal);

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
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          We weren't able to retrieve goals at this time...
        </MSpan>
      </div>
    );
  }

  function ListContent(): JSX.Element {
    if (goals.length === 0) {
      return <EmptyState />;
    }

    return (
      <ul className='w-full flex flex-col gap-2 py-2'>
        { goals
          ?.sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)
          .map(item => (<GoalItem spending={ item } key={ item.spendingId } />)) }
      </ul>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={ SavingsOutlined } title='Goals' >
        <MBaseButton color='primary' className='gap-1 py-1 px-2' onClick={ showNewGoalModal }>
          <AddOutlined />
          New Goal
        </MBaseButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0' ref={ ref }>
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
          <SavingsOutlined className='h-full text-5xl dark:text-dark-monetr-content-muted' />
        </div>
        <MSpan size='xl' color='subtle' className='text-center'>
          You don't have any goals yet...
        </MSpan>
        <MSpan size='lg' color='subtle' className='text-center'>
          Goals are longer budgets that don't recur. They are meant to be used to put money aside for something like
          saving up for a vaction, or paying off a loan.
        </MSpan>
      </div>
    </div>
  );
}
