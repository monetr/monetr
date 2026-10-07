import { isThisYear } from 'date-fns';
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
import { formatRelativeDate } from '@monetr/interface/util/formatDate';

import styles from './TransactionRecurringCard.module.scss';

export interface TransactionRecurringCardProps {
  transaction: Transaction;
}

export default function TransactionRecurringCard(props: TransactionRecurringCardProps): React.JSX.Element | null {
  const { data: recurring } = useRecurringTransaction(props.transaction.transactionRecurringId);
  // The recurring transaction only has the id, look the expense up so it comes from the same cache as the expenses
  const { data: spending } = useSpending(recurring?.spendingId ?? null);
  const { timezone, inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  const { data: localeCurrency } = useLocaleCurrency();

  // The card is just extra detail on top of the transaction, so instead of a skeleton just leave it out until
  // everything has loaded
  if (!recurring || !locale || !localeCurrency) {
    return null;
  }

  const isDebit = recurring.direction === 'debit';

  let eyebrow = 'Recurring Deposit';
  if (isDebit) {
    eyebrow = 'Recurring Charge';
  }

  let title = 'No longer repeating';
  if (!recurring.ended) {
    title = capitalize(rrulestr(recurring.ruleset).toText());
  }

  // Ended ones show when they were last seen, everything else shows when it's expected next
  let dateLabel = 'Next Expected';
  let date = inTimezone(recurring.next);
  if (recurring.ended) {
    dateLabel = 'Last Seen';
    date = inTimezone(recurring.last);
  }

  // TODO These should use formatDate, but right now that formats in the browser's timezone instead of the account's
  const dateString = new Intl.DateTimeFormat(locale.code, {
    month: 'short',
    day: 'numeric',
    // Only include the year when it isn't obvious, otherwise something from last december reads like it's this december
    year: isThisYear(date) ? undefined : 'numeric',
    timeZone: timezone,
  }).format(date);
  const sinceString = new Intl.DateTimeFormat(locale.code, {
    month: 'short',
    year: 'numeric',
    timeZone: timezone,
  }).format(recurring.first);

  return (
    <section className={styles.card} data-testid='transaction-recurring-card'>
      <div className={styles.header}>
        <span className={styles.headerIcon}>
          <Repeat />
        </span>
        <div className={styles.headerText}>
          <span className={styles.eyebrow}>{eyebrow}</span>
          <span className={styles.title}>{title}</span>
        </div>
        <Tooltip delayDuration={100}>
          <TooltipTrigger asChild>
            <span className={styles.confidence}>{recurring.getConfidenceLabel()}</span>
          </TooltipTrigger>
          <TooltipContent side='top'>Confidence {Math.round(recurring.confidence * 100)}%</TooltipContent>
        </Tooltip>
      </div>

      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={styles.statLabel}>{dateLabel}</span>
          <span className={styles.statValue}>{dateString}</span>
          {!recurring.ended && (
            <span className={styles.statDetail}>{formatRelativeDate(recurring.next, inTimezone, locale)}</span>
          )}
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Usually</span>
          <span className={styles.statValue}>
            {localeCurrency.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)}
          </span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Since</span>
          <span className={styles.statValue}>{sinceString}</span>
        </div>
      </div>

      {/* Only money leaving the account can be budgeted for with an expense */}
      {isDebit && recurring.spendingId && (
        <div className={styles.footer}>
          <span className={styles.spentFrom}>
            {recurring.autoMatched && (
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
                  monetr linked this for you since your recent charges were spent from {spending?.name}
                </TooltipContent>
              </Tooltip>
            )}
            {!recurring.autoMatched && (
              <span className={styles.spentFromIcon}>
                <Wallet />
              </span>
            )}
            <span>
              Budgeted with <strong>{spending?.name}</strong>
            </span>
          </span>
          <Button asChild variant='secondary'>
            <Link to={`/bank/${recurring.bankAccountId}/expenses/${recurring.spendingId}/details`}>View Expense</Link>
          </Button>
        </div>
      )}
      {isDebit && !recurring.spendingId && (
        <div className={styles.footer}>
          <span className={styles.spentFrom}>
            <span className={styles.spentFromIcon}>
              <Wallet />
            </span>
            <SpentFrom recurring={recurring} />
          </span>
          {!recurring.ended && (
            <Button
              onClick={() =>
                showNewExpenseModal({
                  recurring: recurring,
                  transaction: props.transaction,
                })
              }
              variant='primary'
            >
              <Plus />
              Create Expense
            </Button>
          )}
        </div>
      )}

      {/* Money coming in, like a paycheck, can be used to fund expenses with a funding schedule */}
      {!isDebit && recurring.fundingSchedule && (
        <div className={styles.footer}>
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
              View Funding Schedule
            </Link>
          </Button>
        </div>
      )}
      {!isDebit && !recurring.fundingSchedule && !recurring.ended && (
        <div className={styles.footer}>
          <span className={styles.spentFrom}>
            <span className={styles.spentFromIcon}>
              <CalendarSync />
            </span>
            <span>Use this deposit to fund your budgets</span>
          </span>
          <Button
            onClick={() =>
              showNewFundingModal({
                recurring: recurring,
                transaction: props.transaction,
              })
            }
            variant='primary'
          >
            <Plus />
            Create Funding Schedule
          </Button>
        </div>
      )}
    </section>
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
