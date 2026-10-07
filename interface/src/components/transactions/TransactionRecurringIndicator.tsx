import { isThisYear } from 'date-fns';
import { Repeat } from 'lucide-react';
import { rrulestr } from 'rrule';

import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';

import styles from './TransactionRecurringIndicator.module.scss';

export interface TransactionRecurringIndicatorProps {
  transactionRecurringId: ID<TransactionRecurring>;
}

export default function TransactionRecurringIndicator(props: TransactionRecurringIndicatorProps): React.JSX.Element {
  return (
    <Tooltip delayDuration={100}>
      <TooltipTrigger asChild>
        <Repeat aria-label='Transaction is recurring' className={styles.indicator} />
      </TooltipTrigger>
      <TooltipContent align='start' side='top'>
        <RecurringTooltipContent transactionRecurringId={props.transactionRecurringId} />
      </TooltipContent>
    </Tooltip>
  );
}

// Only mounted while the tooltip is open, so the recurring transaction is fetched when the user hovers the indicator.
// Touch devices never open the tooltip, the transaction details page covers that case instead.
function RecurringTooltipContent({ transactionRecurringId }: TransactionRecurringIndicatorProps): React.JSX.Element {
  const { data: recurring, isLoading: recurringIsLoading } = useRecurringTransaction(transactionRecurringId);
  const { timezone, inTimezone } = useTimezone();
  const { data: locale, isLoading: localeIsLoading } = useLocale();
  const { data: localeCurrency, isLoading: localeCurrencyIsLoading } = useLocaleCurrency();

  if (recurringIsLoading || localeIsLoading || localeCurrencyIsLoading) {
    return (
      <div aria-label='Loading recurring details' className={styles.skeleton} role='status'>
        <div className={styles.titlePlaceholder} />
        <div className={styles.subtitlePlaceholder} />
      </div>
    );
  }

  if (!recurring || !locale || !localeCurrency) {
    return <span className={styles.title}>Transaction is recurring</span>;
  }

  const amount = (
    <span className={styles.amount} data-positive={recurring.direction === 'credit'}>
      {localeCurrency.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)}
    </span>
  );

  // Ended ones show when they were last seen, everything else shows when it's expected next
  let date = inTimezone(recurring.next);
  if (recurring.ended) {
    date = inTimezone(recurring.last);
  }

  // TODO This should use formatDate, but right now that formats in the browser's timezone instead of the account's
  const dateString = new Intl.DateTimeFormat(locale.code, {
    month: 'short',
    day: 'numeric',
    year: isThisYear(date) ? undefined : 'numeric',
    timeZone: timezone,
  }).format(date);

  if (recurring.ended) {
    return (
      <div className={styles.tooltip}>
        <span className={styles.title}>No longer repeating</span>
        <span className={styles.subtitle}>
          Last one was {dateString}, {amount}
        </span>
      </div>
    );
  }

  const title = capitalize(rrulestr(recurring.ruleset).toText());

  return (
    <div className={styles.tooltip}>
      <span className={styles.title}>{title}</span>
      <span className={styles.subtitle}>
        Next one expected {dateString}, around {amount}
      </span>
    </div>
  );
}
