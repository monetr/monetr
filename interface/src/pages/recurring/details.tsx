import { Fragment, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { differenceInCalendarDays, isThisYear, startOfToday } from 'date-fns';
import { ChevronRight, Clock, HeartCrack, Layers, Plus, Repeat, Sparkles } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link, useRoute } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import { SelectSpendingOptionComponent } from '@monetr/interface/components/MSelectSpending';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { Switch } from '@monetr/interface/components/Switch';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { getConfidenceLabel } from '@monetr/interface/components/transactions/TransactionRecurringCard';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchTransactionRecurring } from '@monetr/interface/hooks/usePatchTransactionRecurring';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useSimilarTransactions } from '@monetr/interface/hooks/useSimilarTransactions';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import { FREE_TO_USE, FreeToUse, type default as Spending, SpendingType } from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

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
            <RecurringLink name={name} recurring={recurring} />
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
  const { data: cluster } = useTransactionCluster(recurring.transactionClusterId);
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
          {cluster?.originalMemo && (
            <span className={styles.showsUpAs}>
              <span className={styles.showsUpAsLabel}>Shows up as</span>
              <span className={styles.memo} title={cluster.originalMemo}>
                {cluster.originalMemo}
              </span>
            </span>
          )}
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

// RecurringLink is where they pick what the charges on this schedule come out of, or what a recurring deposit funds.
function RecurringLink(props: RecurringProps): React.JSX.Element {
  if (props.recurring.direction === 'debit') {
    return <SpendingLink {...props} />;
  }

  return <FundingLink {...props} />;
}

type SpendingOption = Pick<Spending | FreeToUse, 'spendingId' | 'spendingType' | 'currentAmount' | 'name'>;

// SpendingLink follows the design for the schedule rule, picking an expense and a switch for automatically spending new
// charges from it. Only expenses can be linked for now, goals will get their own treatment later.
function SpendingLink({ recurring, name }: RecurringProps): React.JSX.Element {
  const { data: spending, isLoading: spendingIsLoading } = useSpendings();
  const { data: balances, isLoading: balancesIsLoading } = useCurrentBalance();
  const { data: localeCurrency } = useLocaleCurrency();
  const { saving, link } = useLinkRecurring(recurring);

  const options: Array<SelectOption<SpendingOption>> = useMemo(
    () => [
      // free-to-use is how you say it isn't budgeted with anything, same as everywhere else spending gets picked
      ...(balances ? [{ label: 'Free-To-Use', value: new FreeToUse(balances) }] : []),
      ...(spending ?? [])
        .filter(item => item.spendingType === SpendingType.Expense)
        .sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()))
        .map(item => ({ label: item.name, value: item })),
    ],
    [balances, spending],
  );
  // anything past free-to-use is something they could actually pick
  const isEmpty = !options.some(option => option.value.spendingId !== FREE_TO_USE);
  const value = options.find(option => option.value.spendingId === (recurring.spendingId ?? FREE_TO_USE));
  const amount = localeCurrency?.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored);

  function onChange(newValue: SelectOption<SpendingOption>) {
    const spendingId = newValue.value.spendingId === FREE_TO_USE ? null : (newValue.value.spendingId as ID<Spending>);
    if (spendingId === recurring.spendingId) {
      return;
    }

    return link({ spendingId }, spendingId ? `Linked to ${newValue.label}` : 'No longer budgeted with an expense');
  }

  return (
    <section className={styles.link}>
      <div className={styles.linkHeader}>
        <Typography component='h3' size='lg' weight='semibold'>
          Automatically spend
        </Typography>
        <Typography color='subtle' size='sm'>
          Pick where new {name || 'recurring'} charges come out of when they show up.
        </Typography>
      </div>
      <div className={styles.rule} data-active={Boolean(recurring.spending)}>
        <div className={styles.ruleHeader}>
          <div className={styles.ruleText}>
            <span className={styles.ruleChip}>
              <Repeat />
              This schedule
            </span>
            <span className={styles.ruleTitle}>Spend charges on this schedule from</span>
            <span className={styles.ruleDescription}>
              Only the {amount ? `${amount} ` : ''}charges monetr matches to this schedule.
            </span>
          </div>
          {/* automatically spending isnt a thing yet, the switch is just here so its obvious where it will live */}
          <Tooltip delayDuration={100}>
            <TooltipTrigger asChild>
              <span>
                <Switch aria-label='Automatically spend charges on this schedule' checked={false} disabled />
              </span>
            </TooltipTrigger>
            <TooltipContent side='top'>Automatically spending new charges is coming soon.</TooltipContent>
          </Tooltip>
        </div>
        <div className={styles.ruleControls}>
          <Select
            className={styles.ruleSelect}
            disabled={saving || isEmpty}
            isLoading={spendingIsLoading || balancesIsLoading}
            onChange={onChange}
            optionComponent={SelectSpendingOptionComponent}
            options={options}
            placeholder={isEmpty ? 'No expenses exist...' : 'Select an expense...'}
            value={isEmpty ? undefined : value}
          />
          {/* creating one only makes sense when nothing is picked yet */}
          {!recurring.spending && !recurring.ended && (
            <Button
              className={styles.linkButton}
              disabled={saving}
              onClick={() => showNewExpenseModal({ recurring })}
              variant='secondary'
            >
              <Plus />
              New expense
            </Button>
          )}
        </div>
        {recurring.spending && recurring.autoMatched && (
          <span className={styles.linkNote}>
            <Sparkles />
            monetr picked {recurring.spending.name} for you since your recent charges were spent from it.
          </span>
        )}
        {recurring.spending && (
          <Link
            className={styles.linkView}
            to={`/bank/${recurring.bankAccountId}/expenses/${recurring.spending.spendingId}/details`}
          >
            View expense
          </Link>
        )}
      </div>
    </section>
  );
}

// FundingLink is the same idea for money coming in, like a paycheck funding a schedule. Theres nothing to automatically
// spend here so theres no switch.
function FundingLink({ recurring, name }: RecurringProps): React.JSX.Element {
  const { data: funding, isLoading: fundingIsLoading } = useFundingSchedules();
  const { saving, link } = useLinkRecurring(recurring);

  const options: Array<SelectOption<ID<FundingSchedule> | null>> = useMemo(
    () => [
      { label: 'Nothing', value: null },
      ...(funding ?? [])
        .sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()))
        .map(item => ({ label: item.name, value: item.fundingScheduleId })),
    ],
    [funding],
  );
  // options always has the nothing option, so anything past that is something they could actually pick
  const isEmpty = options.length === 1;
  const value = options.find(option => option.value === recurring.fundingScheduleId);

  function onChange(newValue: SelectOption<ID<FundingSchedule> | null>) {
    if (newValue.value === recurring.fundingScheduleId) {
      return;
    }

    return link(
      { fundingScheduleId: newValue.value },
      newValue.value ? `Linked to ${newValue.label}` : 'No longer funding anything',
    );
  }

  return (
    <section className={styles.link}>
      <div className={styles.linkHeader}>
        <Typography component='h3' size='lg' weight='semibold'>
          Funds
        </Typography>
        <Typography color='subtle' size='sm'>
          The funding schedule {name || 'these'} deposits go towards.
        </Typography>
      </div>
      <div className={styles.rule} data-active={Boolean(recurring.fundingSchedule)}>
        <div className={styles.ruleControls}>
          <Select
            className={styles.ruleSelect}
            disabled={saving || isEmpty}
            isLoading={fundingIsLoading}
            onChange={onChange}
            options={options}
            placeholder={isEmpty ? 'No funding schedules exist...' : 'Select a funding schedule...'}
            value={isEmpty ? undefined : value}
          />
          {!recurring.fundingSchedule && !recurring.ended && (
            <Button
              className={styles.linkButton}
              disabled={saving}
              onClick={() => showNewFundingModal({ recurring })}
              variant='secondary'
            >
              <Plus />
              New funding schedule
            </Button>
          )}
        </div>
        {recurring.fundingSchedule && (
          <Link
            className={styles.linkView}
            to={`/bank/${recurring.bankAccountId}/funding/${recurring.fundingSchedule.fundingScheduleId}/details`}
          >
            View funding schedule
          </Link>
        )}
      </div>
    </section>
  );
}

// useLinkRecurring patches the links on the recurring transaction and lets them know how it went. A null still gets
// sent so picking nothing actually clears the link.
function useLinkRecurring(recurring: TransactionRecurring) {
  const patchTransactionRecurring = usePatchTransactionRecurring();
  const { enqueueSnackbar } = useSnackbar();
  const [saving, setSaving] = useState(false);

  async function link(
    links: { spendingId?: ID<Spending> | null; fundingScheduleId?: ID<FundingSchedule> | null },
    message: string,
  ) {
    setSaving(true);
    return await patchTransactionRecurring({
      transactionRecurringId: recurring.transactionRecurringId,
      bankAccountId: recurring.bankAccountId,
      ...links,
    })
      .then(() => void enqueueSnackbar(message, { variant: 'success', disableWindowBlurListener: true }))
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response?.data?.error || 'Failed to update the recurring transaction', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => setSaving(false));
  }

  return { saving, link };
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
  const memos = useClusterMemos(recurring);
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
      {memos.length > 0 && (
        <div className={styles.clusterMemos}>
          <span className={styles.clusterMemosLabel}>Grouped because they show up as</span>
          <div className={styles.clusterMemoList}>
            {memos.map(([memo, count]) => (
              <Fragment key={memo}>
                <span className={styles.memo} title={memo}>
                  {memo}
                </span>
                <span className={styles.clusterMemoCount}>{count === 1 ? '1 charge' : `${count} charges`}</span>
              </Fragment>
            ))}
          </div>
        </div>
      )}
      <span className={styles.clusterNote}>
        Anything else from {name || 'this group'} that isn&apos;t on this schedule shows up as one-off below.
      </span>
    </section>
  );
}

// useClusterMemos gives back each different way the charges in the group showed up on the statement, most common first.
// Its only the most recent 100 charges, which covers years of most things that repeat.
function useClusterMemos(recurring: TransactionRecurring): Array<[string, number]> {
  const { data: transactions } = useQuery<Array<WithJsonValues<Transaction>>>({
    queryKey: [
      'GET',
      `/api/bank_accounts/${recurring.bankAccountId}/transactions`,
      { transaction_cluster_id: recurring.transactionClusterId.toString(), limit: 100 },
    ],
  });

  return useMemo(() => {
    const counts = (transactions ?? []).reduce<Map<string, number>>((accumulator, item) => {
      if (item.originalName) {
        accumulator.set(item.originalName, (accumulator.get(item.originalName) ?? 0) + 1);
      }
      return accumulator;
    }, new Map());
    return Array.from(counts.entries()).sort((a, b) => b[1] - a[1]);
  }, [transactions]);
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

function ExpectedItem({ recurring }: RecurringProps): React.JSX.Element | null {
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
            {formatDate(recurring.next, inTimezone, locale, DateLength.Full)}
          </Typography>
          <Typography color='subtle' component='p' ellipsis size='sm' weight='medium'>
            Expected next
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
          {/* every charge in here is the same merchant, so the date leads and the raw bank text is what tells them apart */}
          <span className={styles.chargeDate}>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Full)}
            {!onSchedule && <span className={styles.oneOffChip}>One-off</span>}
          </span>
          <span className={styles.memo} title={transaction.originalName}>
            {transaction.originalName}
          </span>
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
