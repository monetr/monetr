/* eslint-disable max-len */
import React from 'react';
import { DayPicker, DayPickerRangeProps, DayPickerSingleProps, useNavigation } from 'react-day-picker';
import { ArrowLeftOutlined, ArrowRightOutlined, ChevronLeft, ChevronRight } from '@mui/icons-material';

import { MBaseButton } from './MButton';
import MSpan from './MSpan';

import { addYears, format } from 'date-fns';

export type MCalendarProps<T extends DayPickerSingleProps | DayPickerRangeProps> = T & {
  enableYearNavigation: boolean;
};

export default function MCalendar<T extends DayPickerSingleProps | DayPickerRangeProps>(props: MCalendarProps<T>) {
  const {
    mode,
    defaultMonth,
    selected,
    onSelect,
    locale,
    disabled,
    enableYearNavigation,
    classNames,
    ...remainingProps
  } = props;
  return (
    <DayPicker
      showOutsideDays={ true }
      mode={ mode as any }
      defaultMonth={ defaultMonth }
      selected={ selected }
      onSelect={ onSelect as any }
      locale={ locale }
      disabled={ disabled }
      classNames={ {
        months: 'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0',
        month: 'space-y-4',
        caption: 'flex justify-center pt-2 relative items-center',
        caption_label: 'text-monetr-default text-monetr-content-emphasis dark:text-dark-monetr-content-emphasis font-medium',
        nav: 'space-x-1 flex items-center',
        nav_button: 'flex items-center justify-center p-1 h-7 w-7 outline-none focus:ring-2 transition duration-100 border border-monetr-border dark:border-dark-monetr-border hover:bg-tremor-background-muted dark:hover:bg-dark-tremor-background-muted rounded-tremor-small focus:border-tremor-brand-subtle dark:focus:border-dark-tremor-brand-subtle focus:ring-tremor-brand-muted dark:focus:ring-dark-tremor-brand-muted text-tremor-content-subtle dark:text-dark-tremor-content-subtle hover:text-tremor-content dark:hover:text-dark-tremor-content',
        nav_button_previous: 'absolute left-1',
        nav_button_next: 'absolute right-1',
        table: 'w-full border-collapse space-y-1',
        head_row: 'flex',
        head_cell: 'w-9 font-normal text-center text-monetr-content-subtle dark:text-dark-monetr-content-subtle',
        row: 'flex w-full mt-0.5',
        cell: 'text-center p-0 relative focus-within:relative text-monetr-default text-monetr-content-emphasis dark:text-dark-monetr-content-emphasis',
        day: 'h-9 w-9 p-0 hover:bg-monetr-background-subtle dark:hover:bg-dark-monetr-background-subtle outline-monetr-brand dark:outline-dark-monetr-brand rounded-tremor-default',
        day_today: 'dark:bg-dark-monetr-background-focused dark:text-dark-monetr-brand-muted',
        day_selected: 'dark:aria-selected:bg-dark-monetr-brand dark:aria-selected:text-dark-monetr-content-emphasis',
        day_disabled: 'dark:text-dark-monetr-content-subtle disabled:hover:bg-transparent',
        day_outside: 'dark:text-dark-monetr-content-subtle',
        ...classNames,
      } }
      components={ {
        IconLeft: ({ ...props }) => <ArrowLeftOutlined className="h-4 w-4" { ...props } />,
        IconRight: ({ ...props }) => <ArrowRightOutlined className="h-4 w-4" { ...props } />,
        Caption: ({ ...props }) => {
          const { goToMonth, nextMonth, previousMonth, currentMonth } = useNavigation();

          return (
            <div className="flex justify-between items-center" { ...props }>
              <div className="flex items-center space-x-1">
                {enableYearNavigation && (
                  <MBaseButton
                    onClick={ () => currentMonth && goToMonth(addYears(currentMonth, -1)) }
                  >
                    Left Year
                  </MBaseButton>
                )}
                <MBaseButton
                  variant='text'
                  onClick={ () => previousMonth && goToMonth(previousMonth) }
                >
                  <ChevronLeft className='text-lg' />
                </MBaseButton>
              </div>

              <MSpan>
                {format(props.displayMonth, 'LLLL yyy', { locale })}
              </MSpan>

              <div className="flex items-center space-x-1">
                <MBaseButton
                  variant='text'
                  onClick={ () => nextMonth && goToMonth(nextMonth) }
                >
                  <ChevronRight className='text-lg' />
                </MBaseButton>
                {enableYearNavigation && (
                  <MBaseButton
                    onClick={ () => currentMonth && goToMonth(addYears(currentMonth, 1)) }
                  >
                    Right Year
                  </MBaseButton>
                )}
              </div>
            </div>
          );
        },
      } }
      { ...remainingProps }
    />
  );
}
