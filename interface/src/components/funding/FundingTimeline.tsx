/* eslint-disable max-len */
import React, { useMemo } from 'react';
import NorthEast from '@mui/icons-material/NorthEast';
import { tz } from '@date-fns/tz';
import { format, getUnixTime } from 'date-fns';

import MSpan from '@monetr/interface/components/MSpan';
import { Event, useForecast } from '@monetr/interface/hooks/forecast';
import { useFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

interface FundingTimelineProps {
  fundingScheduleId: string;
}

interface TimelineItemData {
  date: Date;
  originalDate: Date;
  contributedAmount: number;
  totalContributedAmount: number;
  endingAllocation: number;
  spendingCount: number;
}

export default function FundingTimeline(props: FundingTimelineProps): JSX.Element {
  const { data: timezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const { data: funding } = useFundingSchedule(props.fundingScheduleId);
  const { result: forecast, isLoading, isError } = useForecast();
  const inTimezone = useMemo(() => tz(timezone), [timezone]);

  if (isLoading) {
    return (
      <MSpan>Loading...</MSpan>
    );
  }

  if (isError || !funding) {
    return (
      <MSpan>Failed to load funding forecast!</MSpan>
    );
  }

  // Keep only the events that have spending or contributions for this spending object.
  const events: Array<Event> = (forecast?.events || [])
    .filter(event => event.funding.some(funding => funding.fundingScheduleId === props.fundingScheduleId))
    .map(event => ({
      ...event,
      funding: event.funding.filter(funding => funding.fundingScheduleId === props.fundingScheduleId) || [],
    }));

  // Take all of those events and prepare our data for the timeline.
  const timelineItems: Array<TimelineItemData> = events.map(event => {
    const item: TimelineItemData = {
      date: event.date,
      originalDate: event.funding.find(funding => funding.fundingScheduleId === props.fundingScheduleId).originalDate,
      totalContributedAmount: event.contribution,
      contributedAmount: 0,
      endingAllocation: event.balance,
      spendingCount: 0,
    };
  
    event.spending.forEach(spending => {
      if (!spending.funding.find(funding => funding.fundingScheduleId === props.fundingScheduleId)) {
        return;
      }
  
      item.contributedAmount += spending.contributionAmount;
      item.spendingCount++;
    });

    return item;
  });

  function TimelineItem(props: TimelineItemData & { last: boolean }): JSX.Element {
    let header = '';
    let body = '';
    let dateExtra = '';
    const secondaryBody: string | null = null;
    let icon: JSX.Element | null = null;
    // Only contributed
    header = 'Contribution';
    icon = <NorthEast />;
    body = `${funding.name} will contribute ${locale.formatAmount(props.contributedAmount, AmountType.Stored)} to ${props.spendingCount} budget(s), resulting in a total allocation of ${locale.formatAmount(props.endingAllocation, AmountType.Stored)}.`;
    if (funding.estimatedDeposit) {
      body += ' ';
      body += `An estimated ${locale.formatAmount(funding.estimatedDeposit - props.contributedAmount, AmountType.Stored)} will be left over for Free-to-Use after this contribution.`;
    }
    if (props.date.getDate() != props.originalDate.getDate()) {
      dateExtra = `(Avoided weekend or holiday on ${format(inTimezone(props.originalDate), 'MMMM do')})`;
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
        <time className='mb-1 text-sm font-normal leading-none text-zinc-400 dark:text-zinc-500'>
          {format(inTimezone(props.date), 'MMMM do')} <br /> { dateExtra }
        </time>
        <h3 className='text-lg font-semibold text-zinc-900 dark:text-white'>
          {header} {icon}
        </h3>
        <p className='text-base font-normal text-zinc-500 dark:text-zinc-400'>
          {body}
        </p>
        {secondaryBody && <p className='text-base font-normal text-zinc-500 dark:text-zinc-400'>{secondaryBody}</p>}
      </li>
    );
  }

  return (
    <ol className='relative border-l border-zinc-200 dark:border-zinc-700'>
      {timelineItems.map((item, index) => (<TimelineItem key={ getUnixTime(item.date) } { ...item } last={ timelineItems.length - 1 === index } />))}
    </ol>
  );
}
