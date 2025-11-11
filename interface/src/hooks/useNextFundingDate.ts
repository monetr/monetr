import { isBefore } from 'date-fns';

import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { useLocale } from '@monetr/interface/hooks/useLocale';

/**
 *  useNextFundingDate will return a M/DD formatted string showing when the next funding schedule will recur. This is
 *  just the earliest funding shedule among all the funding schedules for the current bank account.
 */
export function useNextFundingDate(): string | null {
  const { data: funding } = useFundingSchedules();
  const { data: locale } = useLocale();
  const date = funding?.sort((a, b) => (isBefore(a.nextRecurrence, b.nextRecurrence) ? 1 : -1)).pop();

  if (date && locale) {
    return new Intl.DateTimeFormat(locale.code, {
      month: 'numeric',
      day: 'numeric',
    }).format(date.nextRecurrence);
  }

  return null;
}
