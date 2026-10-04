import { Fragment } from 'react';
import { differenceInCalendarDays, isThisYear, startOfToday } from 'date-fns';
import { ArrowRight, Plus, Repeat, TrendingDown, TrendingUp, Wallet } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link } from 'wouter';

import { Button } from '@monetr/interface/components/Button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useRecurringTransactionHistory } from '@monetr/interface/hooks/useRecurringTransactionHistory';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import type Transaction from '@monetr/interface/models/Transaction';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';

import styles from './TransactionRecurringCard.module.scss';

export interface TransactionRecurringCardProps {
  transaction: Transaction;
}

export default function TransactionRecurringCard({
  transaction,
}: TransactionRecurringCardProps): React.JSX.Element | null {
  const { data: recurring } = useRecurringTransaction(transaction.transactionRecurringId);
  const { transactions, seen, priceChange } = useRecurringTransactionHistory(recurring);
  const { timezone, inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();

  // the card is just extra detail on top of the transaction, so instead of a skeleton just leave it out until everything
  // has loaded
  if (!recurring || !locale || !localeCurrency) {
    return null;
  }

  const isDebit = recurring.direction === 'debit';
  const formatAmount = (amount: number) => localeCurrency.formatAmount(Math.abs(amount), AmountType.Stored);
  const formatDate = (date: Date, options: Intl.DateTimeFormatOptions) =>
    new Intl.DateTimeFormat(locale.code, { ...options, timeZone: timezone }).format(date);

  const daysUntilNext = differenceInCalendarDays(inTimezone(recurring.next), startOfToday({ in: inTimezone }));
  const PriceChangeIcon =
    priceChange && priceChange.currentAmount > priceChange.previousAmount ? TrendingUp : TrendingDown;

  return (
    <section className={styles.card} data-testid='transaction-recurring-card'>
      <div className={styles.header}>
        <span className={styles.headerIcon}>
          <Repeat />
        </span>
        <div className={styles.headerText}>
          <span className={styles.eyebrow}>{isDebit ? 'Recurring charge' : 'Recurring deposit'}</span>
          <span className={styles.title}>
            {recurring.ended ? 'No longer repeating' : capitalize(rrulestr(recurring.ruleset).toText())}
          </span>
        </div>
        <Tooltip delayDuration={100}>
          <TooltipTrigger asChild>
            <span className={styles.confidence}>{getConfidenceLabel(recurring.confidence)}</span>
          </TooltipTrigger>
          <TooltipContent side='top'>Confidence {recurring.confidence.toFixed(2)}</TooltipContent>
        </Tooltip>
      </div>

      <div className={styles.stats}>
        {recurring.ended ? (
          <Stat label='Last seen' value={formatDate(recurring.last, { month: 'short', day: 'numeric' })} />
        ) : (
          <Stat
            detail={getRelativeDays(daysUntilNext)}
            label='Next expected'
            value={formatDate(recurring.next, { month: 'short', day: 'numeric' })}
          />
        )}
        <Stat label='Usually' value={formatAmount(recurring.lastAmount)} />
        <Stat label='Since' value={formatDate(recurring.first, { month: 'short', year: 'numeric' })} />
        <Stat label='Seen' value={seen === 1 ? '1 time' : `${seen} times`} />
      </div>

      {priceChange && (
        <div className={styles.priceChange}>
          <span className={styles.priceChangeTitle}>
            <PriceChangeIcon />
            Price went {priceChange.currentAmount > priceChange.previousAmount ? 'up' : 'down'} in{' '}
            {formatDate(priceChange.changedAt, {
              month: 'long',
              year: isThisYear(priceChange.changedAt) ? undefined : 'numeric',
            })}
          </span>
          <div className={styles.priceChangeAmounts}>
            <span className={styles.priceChip}>
              {formatAmount(priceChange.previousAmount)} <span>× {priceChange.previousCount}</span>
            </span>
            <ArrowRight />
            <span className={styles.priceChip} data-current>
              {formatAmount(priceChange.currentAmount)} <span>× {priceChange.currentCount}</span>
            </span>
          </div>
        </div>
      )}

      {/* only money leaving the account can be budgeted for with an expense */}
      {isDebit && (
        <div className={styles.footer}>
          {recurring.spending ? (
            <Fragment>
              <span className={styles.spentFrom}>
                <span className={styles.spentFromIcon}>
                  <Wallet />
                </span>
                <span>
                  Budgeted with <strong>{recurring.spending.name}</strong>
                </span>
              </span>
              <Button asChild variant='secondary'>
                <Link
                  to={`/bank/${recurring.spending.bankAccountId}/expenses/${recurring.spending.spendingId}/details`}
                >
                  View expense
                </Link>
              </Button>
            </Fragment>
          ) : (
            <Fragment>
              <span className={styles.spentFrom}>
                <span className={styles.spentFromIcon}>
                  <Wallet />
                </span>
                <SpentFrom seen={seen} transactions={transactions} />
              </span>
              {!recurring.ended && (
                <Button onClick={() => showNewExpenseModal({ recurring, transaction })} variant='primary'>
                  <Plus />
                  Create expense
                </Button>
              )}
            </Fragment>
          )}
        </div>
      )}
    </section>
  );
}

interface StatProps {
  label: string;
  value: string;
  detail?: string;
}

function Stat({ label, value, detail }: StatProps): React.JSX.Element {
  return (
    <div className={styles.stat}>
      <span className={styles.statLabel}>{label}</span>
      <span className={styles.statValue}>{value}</span>
      {detail && <span className={styles.statDetail}>{detail}</span>}
    </div>
  );
}

interface SpentFromProps {
  seen: number;
  transactions: Array<Transaction>;
}

function SpentFrom({ seen, transactions }: SpentFromProps): React.JSX.Element | null {
  const latest = transactions.at(0);
  const { data: spending } = useSpending(latest?.spendingId ?? null);
  if (!latest) {
    return null;
  }

  const name = latest.spendingId ? spending?.name : 'Free-To-Use';
  // can only say it was always spent from the same place if we actually have every transaction loaded
  const always =
    transactions.length === seen && transactions.every(item => item.spendingId === latest.spendingId) && seen > 1;
  if (always) {
    return (
      <span>
        Spent from <strong>{name}</strong> all {seen} times
      </span>
    );
  }

  return (
    <span>
      Last spent from <strong>{name}</strong>
    </span>
  );
}

function getConfidenceLabel(confidence: number): string {
  if (confidence >= 0.9) {
    return 'Very likely';
  }
  if (confidence >= 0.75) {
    return 'Likely';
  }
  return 'Possibly';
}

function getRelativeDays(days: number): string {
  if (days === 0) {
    return 'today';
  }
  if (days === 1) {
    return 'tomorrow';
  }
  if (days > 1) {
    return `in ${days} days`;
  }
  if (days === -1) {
    return 'yesterday';
  }
  return `${Math.abs(days)} days ago`;
}
