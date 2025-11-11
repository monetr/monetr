import { tz, TZDate } from '@date-fns/tz';
import { format, isThisYear, type Locale } from 'date-fns';

export enum DateLength {
  /**
   * Short will render the date as a day and month pair in the format for the specified locale.
   * For example `11/15` for November 15th in `en-US`.
   */
  Short = 'short',
  /**
   * Long will render the date as a combination of month day and year, but only if the year of the date is different
   * than the current year. If the year is the same then a combination of the month and day will be returned with the
   * full name of the month in the specified locale.
   * For example `November 15th` for November 15th in `en-US`.
   */
  Long = 'long',
  /**
   * Full will render the same as long, but will always include the year.
   * For example `November 15th, 2025` for November 15th, 2025 in `en-US`.
   */
  Full = 'full',
}

export function formatDate(
  date: Date | TZDate,
  timezone: string | ReturnType<typeof tz>,
  locale: Locale,
  length: DateLength = DateLength.Full,
): string {
  const inTimezone = typeof timezone === 'function' ? timezone : tz(timezone);
  switch (length) {
    case DateLength.Short:
      return new Intl.DateTimeFormat(locale.code, {
        month: 'numeric',
        day: 'numeric',
      }).format(inTimezone(date));
    case DateLength.Long:
      return new Intl.DateTimeFormat(locale.code, {
        month: 'long',
        day: 'numeric',
        // Only include the year if it is a different year than the current year.
        year: isThisYear(date) ? undefined : 'numeric',
      }).format(inTimezone(date));
    case DateLength.Full:
      return format(
        inTimezone(date),
        locale.formatLong.date({
          width: 'long',
        }),
        {
          locale,
        },
      );
  }
}
