import { Fragment } from 'react';
import { CalendarSync, Plus } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import FundingItem from '@monetr/interface/components/funding/FundingItem';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';

import styles from './funding.module.scss';

export default function Funding(): React.JSX.Element {
  const { isError: fundingIsError, isLoading: fundingIsLoading, data: funding } = useFundingSchedules();

  if (fundingIsLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (fundingIsError) {
    return <Typography size='inherit'>Error...</Typography>;
  }

  function ListContent(): React.JSX.Element {
    if (funding.length === 0) {
      return <EmptyState />;
    }

    return (
      <ul className={styles.list}>
        {funding
          ?.sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
          .map(item => (
            <FundingItem funding={item} key={item.fundingScheduleId} />
          ))}
      </ul>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={CalendarSync} title='Funding Schedules'>
        <Button onClick={showNewFundingModal} variant='primary'>
          <Plus />
          New Funding Schedule
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
          <CalendarSync className={styles.emptyIcon} />
        </div>
        <Typography align='center' color='subtle' size='xl'>
          You don't have any funding schedules yet...
        </Typography>
        <Typography align='center' color='subtle' size='lg'>
          Funding schedules tell monetr when to allocate funds towards your expenses and goals.
        </Typography>
      </div>
    </div>
  );
}
