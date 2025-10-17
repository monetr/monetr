import { format, isBefore } from 'date-fns';

import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';

/**
 *  useNextFundingDate will return a M/DD formatted string showing when the next funding schedule will recur. This is
 *  just the earliest funding shedule among all the funding schedules for the current bank account.
 */
export function useNextFundingDate(): string | null {
  const { data: funding } = useFundingSchedules();
  const date = funding?.sort((a, b) => (isBefore(a.nextRecurrence, b.nextRecurrence) ? 1 : -1)).pop();

  if (date) {
    return format(date.nextRecurrence, 'M/dd');
  }

  return null;
}
