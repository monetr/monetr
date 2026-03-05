import { addYears, format } from 'date-fns';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { DayPicker, type DayPickerSingleProps, useNavigation } from 'react-day-picker';

import { Button, buttonVariants } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Calendar.module.scss';

export type CalendarProps = DayPickerSingleProps & {
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
        caption: styles.captionRow,
        caption_label: styles.captionLabel,
        nav: styles.nav,
        nav_button: buttonVariants({ variant: 'calendar' }),
        nav_button_previous: styles.navButtonPrevious,
        nav_button_next: styles.navButtonNext,
        table: styles.table,
        head_row: styles.headRow,
        head_cell: styles.headCell,
        row: styles.row,
        cell: styles.cell,
        day: styles.day,
        day_range_end: 'day-range-end',
        day_selected: styles.daySelected,
        day_today: styles.dayToday,
        day_outside: styles.dayOutside,
        day_disabled: styles.dayDisabled,
        day_range_middle: styles.dayRangeMiddle,
        day_hidden: styles.dayHidden,
        ...classNames,
      }}
      components={{
        // biome-ignore lint/correctness/noNestedComponentDefinitions: Easier to structure it this way.
        Caption: ({ ...props }) => {
          const { goToMonth, nextMonth, previousMonth, currentMonth } = useNavigation();
          return (
            <div className={styles.caption} {...props}>
              <div className={styles.navGroup}>
                {enableYearNavigation && (
                  <Button onClick={() => goToMonth(addYears(currentMonth, -1))} variant='calendar'>
                    <ChevronsLeft />
                  </Button>
                )}
                <Button onClick={() => goToMonth(previousMonth)} variant='calendar'>
                  <ChevronLeft />
                </Button>
              </div>

              <Typography className={styles.monthLabel} color='emphasis' size='sm'>
                {format(props.displayMonth, 'LLLL yyy', { locale })}
              </Typography>

              <div className={styles.navGroup}>
                <Button onClick={() => goToMonth(nextMonth)} variant='calendar'>
                  <ChevronRight />
                </Button>
                {enableYearNavigation && (
                  <Button onClick={() => goToMonth(addYears(currentMonth, 1))} variant='calendar'>
                    <ChevronsRight />
                  </Button>
                )}
              </div>
            </div>
          );
        },
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
