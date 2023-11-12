/* eslint-disable max-len */
import React from 'react';
import { AirlineStopsOutlined, NorthEast, SouthEast } from '@mui/icons-material';
import { format, getUnixTime } from 'date-fns';

import MSpan from '@monetr/interface/components/MSpan';
import { Event, useForecast } from '@monetr/interface/hooks/forecast';
import { useFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import { useSpending } from '@monetr/interface/hooks/spending';
import { formatAmount } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface ExpenseTimelineProps {
  spendingId: number;
}

interface TimelineItemData {
  date: Date;
  spentAmount: number;
  totalSpentAmount: number;
  contributedAmount: number;
  totalContributedAmount: number;
  endingAllocation: number;
}

export default function ExpenseTimeline(props: ExpenseTimelineProps): JSX.Element {
  const { data: spending } = useSpending(props.spendingId);
  const { data: fundingSchedule } = useFundingSchedule(spending?.fundingScheduleId);
  const { result: forecast, isLoading, isError } = useForecast();

  if (isLoading) {
    return (
      <MSpan>Loading...</MSpan>
    );
  }

  if (isError || !spending) {
    return (
      <MSpan>Failed to load expense forecast!</MSpan>
    );
  }

  // Keep only the events that have spending or contributions for this spending object.
  const events: Array<Event> = forecast.events
    .filter(event => event.spending.some(spending => spending.spendingId === props.spendingId))
    .map(event => ({
      ...event,
      spending: event.spending.filter(spending => spending.spendingId === props.spendingId),
    }));

  // Take all of those events and prepare our data for the timeline.
  const timelineItems: Array<TimelineItemData> = events.map(event => {
    const item: TimelineItemData = {
      date: event.date,
      totalContributedAmount: event.contribution,
      totalSpentAmount: event.transaction,
      spentAmount: 0,
      contributedAmount: 0,
      endingAllocation: 0,
    };

    event.spending.forEach(spending => {
      if (spending.spendingId !== props.spendingId) {
        return;
      }

      item.spentAmount += spending.transactionAmount;
      item.contributedAmount += spending.contributionAmount;
      item.endingAllocation += spending.rollingAllocation;
    });

    return item;
  });

  function TimelineItem(props: TimelineItemData & { last: boolean }): JSX.Element {
    let header = '';
    let body = '';
    let secondaryBody: string | null = null;
    let icon: JSX.Element | null = null;
    if (props.contributedAmount > 0 && props.spentAmount > 0) {
      // Spent and contributed
      header = 'Contribution & Spending';
      icon = <AirlineStopsOutlined />;
      // NOTE To repro this, have your funding schedule land on the same day the item is being spent. For example
      // a funding schedule that is 15th and the last day of the month, landing on september 15th (friday) funding
      // an expense that is spent every friday.
      if (props.endingAllocation > 0) {
        body = `An estimated ${formatAmount(props.spentAmount)} will be spent or be ready to spend, ${formatAmount(props.contributedAmount)} was contributed to this budget at the same time. ${formatAmount(props.endingAllocation)} is left over to use from this budget until the next contribution.`;
      } else {
        body = `An estimated ${formatAmount(props.spentAmount)} will be spent or be ready to spend, included the ${formatAmount(props.contributedAmount)} that was contributed to this budget at the same time to account for the spending.`;
      }
    } else if (props.contributedAmount === 0 && props.spentAmount > 0) {
      // Only spent
      header = 'Spending';
      icon = <SouthEast />;
      body = `An estimated ${formatAmount(props.spentAmount)} will be spent or will be ready to spend, from your ${spending.name} budget.`;
    } else if (props.contributedAmount > 0 && props.spentAmount === 0) {
      // Only contributed
      header = 'Contributition';
      icon = <NorthEast />;
      body = `${formatAmount(props.contributedAmount)} will be allocated towards ${spending.name} from ${fundingSchedule.name}, resulting in a total allocation of ${formatAmount(props.endingAllocation)}.`;
      if (props.totalContributedAmount > props.contributedAmount) {
        secondaryBody = `A total of ${formatAmount(props.totalContributedAmount)} will be contributed to all budgets on this day.`;
      }
    } else {
      // Nothing is happening with this expense on this item.
      return null;
    }
    const rowClassNames = mergeTailwind(
      {
        'mb-5': !props.last,
      },
      'ml-4',
    );
    return (
      <li className={ rowClassNames }>
        <div className='absolute w-3 h-3 bg-zinc-200 rounded-full mt-1.5 -left-1.5 border border-white dark:border-zinc-900 dark:bg-zinc-700' />
        <time className='mb-1 text-sm font-normal leading-none text-zinc-400 dark:text-zinc-500'>{format(props.date, 'MMMM do')}</time>
        <h3 className='text-lg font-semibold text-zinc-900 dark:text-white'>{header} {icon}</h3>
        <p className='text-base font-normal text-zinc-500 dark:text-zinc-400'>{body}</p>
        {secondaryBody && <p className='text-base font-normal text-zinc-500 dark:text-zinc-400'>{secondaryBody}</p>}
      </li>
    );
  }


  return (
    <ol className='relative border-l border-zinc-200 dark:border-zinc-700'>
      <li className='mb-5 ml-4'>
        <div className='absolute w-3 h-3 bg-zinc-200 rounded-full mt-1.5 -left-1.5 border border-white dark:border-zinc-900 dark:bg-zinc-700'></div>
        <time className='mb-1 text-sm font-normal leading-none text-zinc-400 dark:text-zinc-500'>{format(forecast.startingTime, 'MMMM do')}</time>
        <h3 className='text-lg font-semibold text-zinc-900 dark:text-white'>
          {spending.name}
          <span className='bg-blue-100 text-blue-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-blue-900 dark:text-blue-300 ml-3'>Today</span>
        </h3>
        <p className='text-base font-normal text-zinc-500 dark:text-zinc-400'>
          {spending.name} currently has {spending.getCurrentAmountString()} allocated towards it.
        </p>
        <p className='mb-4 text-base font-normal text-zinc-500 dark:text-zinc-400'>
          Below is the timeline for this expense over the next month.
        </p>
      </li>
      {timelineItems.map((item, index) => (<TimelineItem key={ getUnixTime(item.date) } { ...item } last={ timelineItems.length - 1 === index } />))}
    </ol>
  );
}
