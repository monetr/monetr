import { Fragment, useMemo, useState } from 'react';
import { differenceInCalendarDays, isThisYear, startOfToday } from 'date-fns';
import { ChevronRight, Clock, HeartCrack, Layers, Repeat } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link, useRoute } from 'wouter';

import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { getConfidenceLabel } from '@monetr/interface/components/transactions/TransactionRecurringCard';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import { ID } from '@monetr/interface/models/ID';
import type Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './details.module.scss';

type ChargesTab = 'all' | 'schedule' | 'oneoff';

export default function RecurringDetails(): React.JSX.Element | null {
  const [, params] = useRoute<{ bankId: string; transactionRecurringId: string }>(
    '/bank/:bankId/recurring/:transactionRecurringId/details',
  );
  const transactionRecurringId = params?.transactionRecurringId
    ? ID.from<TransactionRecurring, string>(params.transactionRecurringId)
    : null;
  const { data: recurring, isLoading, isError } = useRecurringTransaction(transactionRecurringId);
  const { data: cluster } = useTransactionCluster(recurring?.transactionClusterId ?? null);
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();

  if (!transactionRecurringId) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>There wasn&apos;t a recurring transaction specified...</Typography>
      </div>
    );
  }

  if (isLoading || !locale || !localeCurrency) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError || !recurring) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>Couldn&apos;t find the recurring transaction you specified...</Typography>
      </div>
    );
  }

  const name = cluster?.name ?? '';

  return (
    <Fragment>
      <MTopNavigation
        base={`/bank/${recurring.bankAccountId}/recurring`}
        breadcrumb={name}
        icon={Repeat}
        title='Recurring'
      />
      <div className={styles.body}>
        <div className={styles.columns}>
          <div className={styles.column}>
            <RecurringSummary name={name} recurring={recurring} />
          </div>
          <div className={styles.column}>
            <ClusterSummary name={name} recurring={recurring} />
            <Charges name={name} recurring={recurring} />
          </div>
        </div>
      </div>
    </Fragment>
  );
}

interface RecurringProps {
  recurring: TransactionRecurring;
  name: string;
}

function RecurringSummary({ recurring, name }: RecurringProps): React.JSX.Element | null {
  const { timezone, inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();
  // the next few times we expect it, starting with the one the server already worked out
  const comingUp = useMemo(() => {
    const rule = rrulestr(recurring.ruleset);
    const dates = [recurring.next];
    let last = recurring.next;
    for (let i = 0; i < 2; i++) {
      const after = rule.after(last, false);
      if (!after) {
        break;
      }
      dates.push(after);
      last = after;
    }
    return dates;
  }, [recurring]);

  if (!locale || !localeCurrency) {
    return null;
  }

  const isDebit = recurring.direction === 'debit';
  const formatAmount = (amount: number) => localeCurrency.formatAmount(Math.abs(amount), AmountType.Stored);
  const formatShort = (date: Date, options: Intl.DateTimeFormatOptions = {}) =>
    new Intl.DateTimeFormat(locale.code, {
      month: 'short',
      day: 'numeric',
      // only include the year when it isn't obvious, otherwise something from last december reads like this december
      year: isThisYear(date, { in: inTimezone }) ? undefined : 'numeric',
      timeZone: timezone,
      ...options,
    }).format(date);
  const daysUntilNext = differenceInCalendarDays(inTimezone(recurring.next), startOfToday({ in: inTimezone }));

  return (
    <section className={styles.card}>
      <div className={styles.cardHeader}>
        <div className={styles.cardIcon}>
          <MerchantIcon name={name} />
          <span className={styles.cardIconBadge} title='Recurring'>
            <Repeat />
          </span>
        </div>
        <div className={styles.cardHeaderText}>
          <span className={styles.cardEyebrowRow}>
            <span className={styles.cardEyebrow}>{isDebit ? 'Recurring charge' : 'Recurring deposit'}</span>
            <span className={styles.statusBadge} data-ended={recurring.ended}>
              {recurring.ended ? 'Ended' : 'Active'}
            </span>
          </span>
          <span className={styles.cardTitle}>
            {formatAmount(recurring.lastAmount)} {rrulestr(recurring.ruleset).toText()}
          </span>
        </div>
      </div>
      <div className={styles.stats}>
        {recurring.ended ? (
          <Stat label='Last seen' value={formatShort(recurring.last)} />
        ) : (
          <Stat
            detail={new Intl.RelativeTimeFormat(locale.code, { numeric: 'auto' }).format(daysUntilNext, 'day')}
            label='Next expected'
            value={formatShort(recurring.next)}
          />
        )}
        <Stat label={isDebit ? 'Last charge' : 'Last deposit'} value={formatShort(recurring.last)} />
        <Stat label='Since' value={formatShort(recurring.first, { day: undefined, year: 'numeric' })} />
        <Stat
          label='Confidence'
          title={`${Math.round(recurring.confidence * 100)}%`}
          value={getConfidenceLabel(recurring.confidence)}
        />
      </div>
      {!recurring.ended && (
        <div className={styles.comingUp}>
          <span className={styles.statLabel}>Coming up</span>
          <div className={styles.comingUpDates}>
            {comingUp.map((date, index) => (
              <span className={styles.comingUpDate} data-next={index === 0} key={date.toISOString()}>
                {formatShort(date)}
              </span>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}

interface StatProps {
  label: string;
  value: string;
  detail?: string;
  title?: string;
}

function Stat({ label, value, detail, title }: StatProps): React.JSX.Element {
  return (
    <div className={styles.stat}>
      <span className={styles.statLabel}>{label}</span>
      <span className={styles.statValue} title={title}>
        {value}
      </span>
      {detail && <span className={styles.statDetail}>{detail}</span>}
    </div>
  );
}

function ClusterSummary({ recurring, name }: RecurringProps): React.JSX.Element | null {
  const { seen } = useRecurringTransactionHistory(recurring);
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  if (!locale) {
    return null;
  }

  const since = new Intl.DateTimeFormat(locale.code, { month: 'short', year: 'numeric' }).format(
    inTimezone(recurring.first),
  );

  return (
    <section className={styles.cluster}>
      <span className={styles.clusterEyebrow}>
        <Layers />
        Similar transactions group
      </span>
      <span className={styles.clusterTitle}>
        {name} <span>· since {since}</span>
      </span>
      <div className={styles.comingUpDates}>
        <span className={styles.clusterChip}>
          <Repeat />
          {seen} on this schedule
        </span>
      </div>
      <span className={styles.clusterNote}>
        Anything else from {name || 'this group'} that isn&apos;t on this schedule shows up as one-off below.
      </span>
    </section>
  );
}

function Charges({ recurring, name }: RecurringProps): React.JSX.Element {
  const [tab, setTab] = useState<ChargesTab>('all');
  const history = useRecurringTransactionHistory(recurring);
  const similar = useSimilarTransactions(recurring.transactionClusterId.toString());

  let transactions: Array<Transaction>;
  switch (tab) {
    case 'schedule':
      transactions = history.transactions;
      break;
    case 'oneoff':
      transactions = (similar.data ?? []).filter(
        item => item.transactionRecurringId?.toString() !== recurring.transactionRecurringId.toString(),
      );
      break;
    default:
      transactions = similar.data ?? [];
  }
  // the schedule tab already has everything loaded, the other two come from the whole group a page at a time
  const canLoadMore = tab !== 'schedule' && similar.hasNextPage;
  const isLoading = tab === 'schedule' ? history.isLoading : similar.isLoading;

  return (
    <section className={styles.charges}>
      <div className={styles.chargesHeader}>
        <Typography component='h3' size='xl' weight='semibold'>
          {recurring.direction === 'debit' ? 'Charges' : 'Deposits'}
        </Typography>
        <div className={styles.segmented} role='tablist'>
          <SegmentButton onClick={() => setTab('all')} selected={tab === 'all'}>
            All
          </SegmentButton>
          <SegmentButton onClick={() => setTab('schedule')} selected={tab === 'schedule'}>
            On schedule {history.seen}
          </SegmentButton>
          <SegmentButton onClick={() => setTab('oneoff')} selected={tab === 'oneoff'}>
            One-off
          </SegmentButton>
        </div>
      </div>
      <ul className={styles.chargesList}>
        {tab !== 'oneoff' && <ExpectedItem name={name} recurring={recurring} />}
        {transactions.map(transaction => (
          <ChargeItem
            key={transaction.transactionId}
            onSchedule={transaction.transactionRecurringId?.toString() === recurring.transactionRecurringId.toString()}
            recurringId={recurring.transactionRecurringId}
            transaction={transaction}
          />
        ))}
      </ul>
      {!isLoading && transactions.length === 0 && !canLoadMore && (
        <Typography className={styles.chargesEmpty} color='subtle' size='sm'>
          Nothing here.
        </Typography>
      )}
      {canLoadMore && (
        <button
          className={styles.showMore}
          disabled={similar.isFetchingNextPage}
          onClick={() => void similar.fetchNextPage()}
          type='button'
        >
          {similar.isFetchingNextPage ? 'Loading...' : 'Show more'}
        </button>
      )}
    </section>
  );
}

interface SegmentButtonProps {
  children: React.ReactNode;
  selected: boolean;
  onClick: () => void;
}

function SegmentButton({ children, selected, onClick }: SegmentButtonProps): React.JSX.Element {
  return (
    <button aria-selected={selected} className={styles.segment} onClick={onClick} role='tab' type='button'>
      {children}
    </button>
  );
}

function ExpectedItem({ recurring, name }: RecurringProps): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();
  if (recurring.ended || !locale || !localeCurrency) {
    return null;
  }

  return (
    <Item className={styles.expected}>
      <div className={flexVariants({ orientation: 'row', align: 'center' })}>
        <div className={styles.expectedIcon}>
          <Clock />
        </div>
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <Typography component='p' ellipsis size='md' weight='semibold'>
            {name}
          </Typography>
          <Typography color='subtle' component='p' ellipsis size='sm' weight='medium'>
            Expected {formatDate(recurring.next, inTimezone, locale, DateLength.Long)}
          </Typography>
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          <Typography color='subtle' weight='semibold'>
            ~{localeCurrency.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)}
          </Typography>
          {/* theres nothing to link to, but keep the arrows space so the amount lines up with the real rows */}
          <Typography aria-hidden className={styles.expectedArrowSpacer}>
            <ChevronRight />
          </Typography>
        </ItemContent>
      </div>
    </Item>
  );
}

interface ChargeItemProps {
  transaction: Transaction;
  onSchedule: boolean;
  recurringId: TransactionRecurring['transactionRecurringId'];
}

function ChargeItem({ transaction, onSchedule, recurringId }: ChargeItemProps): React.JSX.Element | null {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  if (!locale) {
    return null;
  }

  return (
    <Item>
      <Link
        className={flexVariants({ orientation: 'row', align: 'center' })}
        to={`/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`}
      >
        {/* only call out the ones on this schedule, anything else in the group is a one-off as far as this page cares */}
        <TransactionMerchantIcon
          name={transaction.getName()}
          pending={transaction.isPending}
          transactionRecurringId={onSchedule ? recurringId : null}
        />
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <Typography color='emphasis' component='p' ellipsis size='md' weight='semibold'>
            {transaction.getName()}
          </Typography>
          <Typography color='subtle' component='p' ellipsis size='sm' weight='medium'>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Full)}
            {!onSchedule && ' · One-off'}
          </Typography>
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          <TransactionAmount transaction={transaction} />
          <Typography>
            <ChevronRight />
          </Typography>
        </ItemContent>
      </Link>
    </Item>
  );
}
