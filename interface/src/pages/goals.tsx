import { Fragment } from 'react';
import { HeartCrack, PiggyBank, Plus } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import GoalItem from '@monetr/interface/components/goals/GoalItem';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useSpendingFiltered } from '@monetr/interface/hooks/useSpendingFiltered';
import { showNewGoalModal } from '@monetr/interface/modals/NewGoalModal';
import { SpendingType } from '@monetr/interface/models/Spending';

import styles from './goals.module.scss';

export default function Goals(): React.JSX.Element {
  const { data: goals, isError, isLoading } = useSpendingFiltered(SpendingType.Goal);

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
        <Typography size='2xl'>We weren't able to retrieve goals at this time...</Typography>
      </div>
    );
  }

  function ListContent(): React.JSX.Element {
    if (goals.length === 0) {
      return <EmptyState />;
    }

    return (
      <ul className={styles.list}>
        {goals
          ?.sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
          .map(item => (
            <GoalItem key={item.spendingId} spending={item} />
          ))}
      </ul>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={PiggyBank} title='Goals'>
        <Button onClick={showNewGoalModal} variant='primary'>
          <Plus />
          New Goal
        </Button>
      </MTopNavigation>
      <div className={styles.content}>
        <ListContent />
      </div>
    </Fragment>
  );
}

function EmptyState(): React.JSX.Element {
  return (
    <div className={styles.empty}>
      <div className={styles.emptyInner}>
        <div className={styles.iconRow}>
          <PiggyBank className={styles.emptyIcon} />
        </div>
        <Typography align='center' color='subtle' size='xl'>
          You don't have any goals yet...
        </Typography>
        <Typography align='center' color='subtle' size='lg'>
          Goals are longer budgets that don't recur. They are meant to be used to put money aside for something like
          saving up for a vaction, or paying off a loan.
        </Typography>
      </div>
    </div>
  );
}
