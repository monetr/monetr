import { isBefore } from 'date-fns';

import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

/**
 *  useNextFundingDate will return a M/DD formatted string showing when the next funding schedule will recur. This is
 *  just the earliest funding shedule among all the funding schedules for the current bank account.
 */
export function useNextFundingDate(): string | null {
  const { data: funding } = useFundingSchedules();
  const { data: locale } = useLocale();
  const { inTimezone } = useTimezone();
  const date = funding?.sort((a, b) => (isBefore(a.nextRecurrence, b.nextRecurrence) ? 1 : -1)).pop();

  if (date && locale) {
    return formatDate(date.nextRecurrence, inTimezone, locale, DateLength.Short);
  }

  return null;
}
