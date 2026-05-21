import { format, getUnixTime } from 'date-fns';
import { ArrowUpRight } from 'lucide-react';

import Typography from '@monetr/interface/components/Typography';
import { type ForecastEvent, useForecast } from '@monetr/interface/hooks/useForecast';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './FundingTimeline.module.scss';

interface FundingTimelineProps {
  fundingScheduleId: string;
}

interface TimelineItemData {
  date: Date;
  funding: FundingSchedule;
  originalDate: Date;
  contributedAmount: number;
  totalContributedAmount: number;
  endingAllocation: number;
  spendingCount: number;
}

export default function FundingTimeline(props: FundingTimelineProps): JSX.Element {
  const { data: funding } = useFundingSchedule(props.fundingScheduleId);
  const { data: forecast, isLoading, isError } = useForecast();

  if (isLoading) {
    return <Typography size='inherit'>Loading...</Typography>;
  }

  if (isError || !funding) {
    return <Typography size='inherit'>Failed to load funding forecast!</Typography>;
  }

  // Keep only the events that have spending or contributions for this spending object.
  const events: Array<ForecastEvent> = (forecast?.events || [])
    .filter(event => event.funding.some(funding => funding.fundingScheduleId === props.fundingScheduleId))
    .map(event => ({
      ...event,
      funding: event.funding.filter(funding => funding.fundingScheduleId === props.fundingScheduleId) || [],
    }));

  // Take all of those events and prepare our data for the timeline.
  const timelineItems: Array<TimelineItemData> = events.map(event => {
    const item: TimelineItemData = {
      date: event.date,
      funding,
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

  return (
    <ol className={styles.timeline}>
      {timelineItems.map(item => (
        <TimelineItem key={getUnixTime(item.date)} {...item} />
      ))}
    </ol>
  );
}

function TimelineItem({ funding, ...props }: TimelineItemData): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: currency } = useLocaleCurrency();

  let header = '';
  let body = '';
  let dateExtra = '';
  const secondaryBody: string | null = null;
  let icon: JSX.Element | null = null;
  // Only contributed
  header = 'Contribution';
  icon = <ArrowUpRight className={styles.icon} />;
  body = `${funding.name} will contribute ${currency.formatAmount(props.contributedAmount, AmountType.Stored)} to ${props.spendingCount} budget(s), resulting in a total allocation of ${currency.formatAmount(props.endingAllocation, AmountType.Stored)}.`;
  if (funding.estimatedDeposit) {
    body += ' ';
    body += `An estimated ${currency.formatAmount(funding.estimatedDeposit - props.contributedAmount, AmountType.Stored)} will be left over for Free-to-Use after this contribution.`;
  }
  if (props.date.getDate() !== props.originalDate.getDate() && funding.excludeWeekends) {
    dateExtra = `(Avoided weekend or holiday on ${format(inTimezone(props.originalDate), 'MMMM do')})`;
  }

  return (
    <li className={styles.row}>
      <div className={styles.dot} />
      <time className={styles.date}>
        {format(inTimezone(props.date), 'MMMM do')} <br /> {dateExtra}
      </time>
      <h3 className={styles.header}>
        {header} {icon}
      </h3>
      <p className={styles.body}>{body}</p>
      {secondaryBody && <p className={styles.body}>{secondaryBody}</p>}
    </li>
  );
}
