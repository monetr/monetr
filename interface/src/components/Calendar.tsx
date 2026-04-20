import { addYears, format, type Locale } from 'date-fns';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { DayPicker, type MonthCaptionProps, type PropsBase, type PropsSingle, useDayPicker } from 'react-day-picker';

import { Button, buttonVariants } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Calendar.module.scss';

import { Fragment } from 'react/jsx-runtime';

export type CalendarProps = PropsBase &
  PropsSingle & {
    enableYearNavigation?: boolean;
  };

function Calendar({
  className,
  classNames,
  locale,
  showOutsideDays = true,
  enableYearNavigation = true,
  ...props
}: CalendarProps) {
  return (
    <DayPicker
      className={mergeTailwind(styles.root, className)}
      classNames={{
        months: styles.months,
        month: styles.month,
        month_caption: styles.captionRow,
        nav: styles.nav,
        button_previous: mergeTailwind(buttonVariants({ variant: 'calendar' }), styles.navButtonPrevious),
        button_next: mergeTailwind(buttonVariants({ variant: 'calendar' }), styles.navButtonNext),
        month_grid: styles.table,
        weekdays: styles.headRow,
        weekday: styles.headCell,
        week: styles.row,
        day: styles.cell,
        day_button: styles.day,
        selected: styles.daySelected,
        today: styles.dayToday,
        outside: styles.dayOutside,
        disabled: styles.dayDisabled,
        hidden: styles.dayHidden,
        ...classNames,
      }}
      components={{
        // biome-ignore lint/correctness/noNestedComponentDefinitions: Easier to structure it this way.
        MonthCaption: (captionProps: MonthCaptionProps) => {
          const { goToMonth, nextMonth, previousMonth } = useDayPicker();
          const displayMonth = captionProps.calendarMonth.date;
          return (
            <div className={styles.caption}>
              <div className={styles.navGroup}>
                {enableYearNavigation && (
                  <Button onClick={() => goToMonth(addYears(displayMonth, -1))} variant='calendar'>
                    <ChevronsLeft />
                  </Button>
                )}
                <Button
                  disabled={!previousMonth}
                  onClick={() => previousMonth && goToMonth(previousMonth)}
                  variant='calendar'
                >
                  <ChevronLeft />
                </Button>
              </div>

              <Typography className={styles.monthLabel} color='emphasis' size='sm'>
                {format(displayMonth, 'LLLL yyy', { locale: locale as Locale })}
              </Typography>

              <div className={styles.navGroup}>
                <Button disabled={!nextMonth} onClick={() => nextMonth && goToMonth(nextMonth)} variant='calendar'>
                  <ChevronRight />
                </Button>
                {enableYearNavigation && (
                  <Button onClick={() => goToMonth(addYears(displayMonth, 1))} variant='calendar'>
                    <ChevronsRight />
                  </Button>
                )}
              </div>
            </div>
          );
        },
        // biome-ignore lint/correctness/noNestedComponentDefinitions: Easier to structure it this way.
        PreviousMonthButton: () => <Fragment />,
        // biome-ignore lint/correctness/noNestedComponentDefinitions: Easier to structure it this way.
        NextMonthButton: () => <Fragment />,
        Nav: () => <Fragment />,
      }}
      locale={locale}
      mode='single'
      showOutsideDays={showOutsideDays}
      {...props}
    />
  );
}
Calendar.displayName = 'Calendar';

export { Calendar };
