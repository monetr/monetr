import { Fragment } from 'react';
import { differenceInCalendarDays, isThisYear, startOfToday } from 'date-fns';
import { CalendarSync, Plus, Repeat, Sparkles, Wallet } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link } from 'wouter';

import { Button } from '@monetr/interface/components/Button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useSpending } from '@monetr/interface/hooks/useSpending';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionsForRecurring } from '@monetr/interface/hooks/useTransactionsForRecurring';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import type Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
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
  // only include the year when it isn't obvious, otherwise something from last december reads like it's this december
  // numeric auto gives us today, tomorrow and yesterday instead of in 0 days, in 1 day and 1 day ago
  const formatRelativeDays = (days: number) =>
    new Intl.RelativeTimeFormat(locale.code, { numeric: 'auto' }).format(days, 'day');
  const yearIfNeeded = (date: Date) => (isThisYear(date, { in: inTimezone }) ? undefined : 'numeric');

  const daysUntilNext = differenceInCalendarDays(inTimezone(recurring.next), startOfToday({ in: inTimezone }));

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
          <TooltipContent side='top'>Confidence {Math.round(recurring.confidence * 100)}%</TooltipContent>
        </Tooltip>
      </div>

      <div className={styles.stats}>
        {recurring.ended ? (
          <Stat
            label='Last seen'
            value={formatDate(recurring.last, { month: 'short', day: 'numeric', year: yearIfNeeded(recurring.last) })}
          />
        ) : (
          <Stat
            detail={formatRelativeDays(daysUntilNext)}
            label='Next expected'
            value={formatDate(recurring.next, { month: 'short', day: 'numeric', year: yearIfNeeded(recurring.next) })}
          />
        )}
        <Stat label='Usually' value={formatAmount(recurring.lastAmount)} />
        <Stat label='Since' value={formatDate(recurring.first, { month: 'short', year: 'numeric' })} />
      </div>

      {/* only money leaving the account can be budgeted for with an expense */}
      {isDebit && (
        <div className={styles.footer}>
          {recurring.spending ? (
            <Fragment>
              <span className={styles.spentFrom}>
                {recurring.autoMatched ? (
                  <Tooltip delayDuration={100}>
                    <TooltipTrigger asChild>
                      <span
                        aria-label='Linked automatically'
                        className={styles.spentFromIcon}
                        data-testid='transaction-recurring-auto-matched'
                        role='img'
                      >
                        <Sparkles />
                      </span>
                    </TooltipTrigger>
                    <TooltipContent side='top'>
                      monetr linked this for you since your recent charges were spent from {recurring.spending.name}
                    </TooltipContent>
                  </Tooltip>
                ) : (
                  <span className={styles.spentFromIcon}>
                    <Wallet />
                  </span>
                )}
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
                <SpentFrom recurring={recurring} />
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

      {/* money coming in, like a paycheck, can be used to fund expenses with a funding schedule */}
      {!isDebit && (recurring.fundingSchedule || !recurring.ended) && (
        <div className={styles.footer}>
          {recurring.fundingSchedule ? (
            <Fragment>
              <span className={styles.spentFrom}>
                <span className={styles.spentFromIcon}>
                  <CalendarSync />
                </span>
                <span>
                  Funds <strong>{recurring.fundingSchedule.name}</strong>
                </span>
              </span>
              <Button asChild variant='secondary'>
                <Link
                  to={`/bank/${recurring.fundingSchedule.bankAccountId}/funding/${recurring.fundingSchedule.fundingScheduleId}/details`}
                >
                  View funding schedule
                </Link>
              </Button>
            </Fragment>
          ) : (
            <Fragment>
              <span className={styles.spentFrom}>
                <span className={styles.spentFromIcon}>
                  <CalendarSync />
                </span>
                <span>Use this deposit to fund your budgets</span>
              </span>
              <Button onClick={() => showNewFundingModal({ recurring, transaction })} variant='primary'>
                <Plus />
                Create funding schedule
              </Button>
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
  recurring: TransactionRecurring;
}

function SpentFrom({ recurring }: SpentFromProps): React.JSX.Element | null {
  // Only the most recent one matters here so that's all we ask for
  const { data: transactions } = useTransactionsForRecurring(recurring.transactionRecurringId, 1);
  const latest = transactions?.at(0);
  const { data: spending } = useSpending(latest?.spendingId ?? null);
  if (!latest) {
    return null;
  }

  const name = latest.spendingId ? spending?.name : 'Free-To-Use';
  return (
    <span>
      Last spent from <strong>{name}</strong>
    </span>
  );
}

export function getConfidenceLabel(confidence: number): string {
  if (confidence >= 0.9) {
    return 'Very likely';
  }
  if (confidence >= 0.75) {
    return 'Likely';
  }
  return 'Possibly';
}
