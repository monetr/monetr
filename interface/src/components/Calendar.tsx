import { addYears, format } from 'date-fns';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { DayPicker, type DayPickerSingleProps, useNavigation } from 'react-day-picker';

import { Button, buttonVariants } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

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
      className={mergeTailwind('p-3', className)}
      classNames={{
        months: 'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0',
        month: 'space-y-4',
        caption: 'flex justify-center pt-1 relative items-center',
        caption_label: 'text-sm font-medium',
        nav: 'space-x-1 flex items-center',
        nav_button: buttonVariants({ variant: 'calendar' }),
        nav_button_previous: 'absolute left-1',
        nav_button_next: 'absolute right-1',
        table: 'w-full border-collapse space-y-1',
        head_row: 'flex',
        head_cell: 'text-muted-foreground rounded-md w-9 font-normal text-[0.8rem]',
        row: 'flex w-full mt-2',
        cell: mergeTailwind(
          'h-9 w-9',
          'text-center text-sm',
          'p-0 relative',
          '[&:has([aria-selected].day-range-end)]:rounded-r-md',
          'first:[&:has([aria-selected])]:rounded-l-md',
          'last:[&:has([aria-selected])]:rounded-r-md',
          'focus-within:relative focus-within:z-20',
        ),
        day: mergeTailwind(buttonVariants({ variant: 'text' }), 'h-9 w-9 p-0 font-normal aria-selected:opacity-100'),
        day_range_end: 'day-range-end',
        day_selected: mergeTailwind(
          'bg-dark-monetr-brand text-dark-monetr-content-emphasis',
          'enabled:hover:bg-dark-monetr-brand-subtle hover:text-dark-monetr-content-emphasis',
          'focus:bg-primary focus:text-dark-monetr-content-emphasis',
        ),
        day_today: 'bg-dark-monetr-background-focused text-dark-monetr-brand-muted hover:text-dark-monetr-brand-bright',
        day_outside: mergeTailwind(
          'day-outside',
          'text-dark-monetr-content-subtle',
          'aria-selected:text-dark-monetr-content',
          'aria-selected:bg-dark-monetr-brand/50',
        ),
        day_disabled: 'text-dark-monetr-content-muted',
        day_range_middle: 'aria-selected:bg-accent aria-selected:text-accent-foreground',
        day_hidden: 'invisible',
        ...classNames,
      }}
      components={{
        Caption: ({ ...props }) => {
          const { goToMonth, nextMonth, previousMonth, currentMonth } = useNavigation();
          return (
            <div className='flex justify-between items-center' {...props}>
              <div className='flex items-center space-x-1'>
                {enableYearNavigation && (
                  <Button onClick={() => goToMonth(addYears(currentMonth, -1))} variant='calendar'>
                    <ChevronsLeft />
                  </Button>
                )}
                <Button onClick={() => goToMonth(previousMonth)} variant='calendar'>
                  <ChevronLeft />
                </Button>
              </div>

              <Typography className='mx-1' color='emphasis' size='sm'>
                {format(props.displayMonth, 'LLLL yyy', { locale })}
              </Typography>

              <div className='flex items-center space-x-1'>
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
