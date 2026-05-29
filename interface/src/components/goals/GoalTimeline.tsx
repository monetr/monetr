import { format, getUnixTime } from 'date-fns';
import { ArrowUpRight, CircleCheck, TrendingUpDown } from 'lucide-react';

import Typography from '@monetr/interface/components/Typography';
import { type ForecastEvent, useForecast } from '@monetr/interface/hooks/useForecast';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type Spending from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './GoalTimeline.module.scss';

export interface GoalTimelineProps {
  spendingId: string;
}

interface TimelineItemData {
  date: Date;
  spending: Spending;
  fundingSchedule: FundingSchedule;
  spentAmount: number;
  totalSpentAmount: number;
  contributedAmount: number;
  totalContributedAmount: number;
  endingAllocation: number;
}

export default function GoalTimeline(props: GoalTimelineProps): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const { data: spending } = useSpending(props.spendingId);
  const { data: fundingSchedule } = useFundingSchedule(spending?.fundingScheduleId);
  const { data: forecast, isLoading, isError } = useForecast();

  if (isLoading) {
    return <Typography size='inherit'>Loading...</Typography>;
  }

  if (isError || !spending) {
    return <Typography size='inherit'>Failed to load goal forecast!</Typography>;
  }

  // Keep only the events that have spending or contributions for this spending object.
  const events: Array<ForecastEvent> = forecast.events
    .filter(event => event.spending.some(spending => spending.spendingId === props.spendingId))
    .map(event => ({
      ...event,
      spending: event.spending.filter(spending => spending.spendingId === props.spendingId),
    }));

  // Take all of those events and prepare our data for the timeline.
  const timelineItems: Array<TimelineItemData> = events.map(event => {
    const item: TimelineItemData = {
      date: event.date,
      spending,
      fundingSchedule,
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

  return (
    <ol className={styles.timeline}>
      <li className={styles.firstRow}>
        <div className={styles.dot}></div>
        <time className={styles.date}>{format(inTimezone(forecast.startingTime), 'MMMM do')}</time>
        <h3 className={styles.header}>
          {spending.name}
          <span className={styles.todayBadge}>Today</span>
        </h3>
        <p className={styles.body}>
          {spending.name} currently has {locale.formatAmount(spending.currentAmount, AmountType.Stored)} allocated
          towards it.
        </p>
        <p className={styles.bodyMarginBottom}>Below is the timeline for this goal over the next month.</p>
      </li>
      {timelineItems.map(item => (
        <TimelineItem key={getUnixTime(item.date)} {...item} />
      ))}
    </ol>
  );
}

function TimelineItem({ spending, fundingSchedule, ...props }: TimelineItemData): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();

  let header = '';
  let body = '';
  let secondaryBody: string | null = null;
  let icon: JSX.Element | null = null;
  if (props.contributedAmount > 0 && props.spentAmount > 0) {
    // Spent and contributed
    header = 'Contribution & Completion';
    icon = <TrendingUpDown className={styles.icon} />;
    // NOTE To repro this, have your funding schedule land on the same day the item is being spent. For example
    // a funding schedule that is 15th and the last day of the month, landing on september 15th (friday) funding
    // an expense that is spent every friday.
    if (props.endingAllocation > 0) {
      body = `An estimated ${locale.formatAmount(props.spentAmount, AmountType.Stored)} will be spent or be ready to spend, ${locale.formatAmount(props.contributedAmount, AmountType.Stored)} was contributed to this goal at the same time. ${locale.formatAmount(props.endingAllocation, AmountType.Stored)} is left over to use from this goal until the next contribution.`;
    } else {
      body = `An estimated ${locale.formatAmount(props.spentAmount, AmountType.Stored)} will be spent or be ready to spend, included the ${locale.formatAmount(props.contributedAmount, AmountType.Stored)} that was contributed to this goal at the same time to account for the spending.`;
    }
  } else if (props.contributedAmount === 0 && props.spentAmount > 0) {
    // Only spent
    header = 'Complete';
    icon = <CircleCheck className={styles.icon} />;
    body = `An estimated ${locale.formatAmount(props.spentAmount, AmountType.Stored)} will be spent or will be ready to spend, from your ${spending.name} goal.`;
  } else if (props.contributedAmount > 0 && props.spentAmount === 0) {
    // Only contributed
    header = 'Contribution';
    icon = <ArrowUpRight className={styles.icon} />;
    body = `${locale.formatAmount(props.contributedAmount, AmountType.Stored)} will be allocated towards ${spending.name} from ${fundingSchedule.name}, resulting in a total allocation of ${locale.formatAmount(props.endingAllocation, AmountType.Stored)}.`;
    if (props.totalContributedAmount > props.contributedAmount) {
      secondaryBody = `A total of ${locale.formatAmount(props.totalContributedAmount, AmountType.Stored)} will be contributed to all budgets on this day.`;
    }
  } else {
    // Nothing is happening with this expense on this item.
    return null;
  }
  return (
    <li className={styles.row}>
      <div className={styles.dot} />
      <time className={styles.date}>{format(inTimezone(props.date), 'MMMM do')}</time>
      <h3 className={styles.header}>
        {header} {icon}
      </h3>
      <p className={styles.body}>{body}</p>
      {secondaryBody && <p className={styles.body}>{secondaryBody}</p>}
    </li>
  );
}
