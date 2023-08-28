/* eslint-disable max-len */

import React, { useCallback, useMemo, useState } from 'react';
import { DayPickerSingleProps } from 'react-day-picker';
import { Popover } from '@headlessui/react';
import { CloseOutlined, TodayOutlined } from '@mui/icons-material';
import { useFormikContext } from 'formik';

import MCalendar from './MCalendar';
import MLabel, { MLabelDecorator } from './MLabel';
import { ReactElement } from './types';

import { isEqual, startOfMonth, startOfToday } from 'date-fns';
import enUS from 'date-fns/locale/en-US';
import mergeTailwind from 'util/mergeTailwind';

export interface MDatePickerProps extends
  Omit<React.HTMLAttributes<HTMLDivElement>, 'value' | 'defaultValue'>
{
  value?: Date;
  min?: Date;
  max?: Date;
  children?: ReactElement;
  disabled?: boolean;
  enableClear?: boolean;
  enableYearNavigation?: boolean;
  error?: string;
  label?: string;
  labelDecorator?: MLabelDecorator;
  name?: string;
  placeholder?: string;
  required?: boolean;
}

const MDatePicker = React.forwardRef<HTMLDivElement, MDatePickerProps>((props, ref) => {
  const today = startOfToday();
  const formikContext = useFormikContext();

  const {
    value = formikContext?.values[props.name],
    min: minDate,
    max: maxDate,
    placeholder = 'Select date',
    disabled = false,
    enableClear = false,
    className,
    enableYearNavigation = false,
    ...other
  } = props;

  const [selectedValue, setSelectedValue] = useState<Date | null>(value);

  const disabledDays = useMemo(() => {
    const disabledDays = [];
    if (minDate) disabledDays.push({ before: minDate });
    if (maxDate) disabledDays.push({ after: maxDate });
    return disabledDays;
  }, [minDate, maxDate]);

  const hasValue = Boolean(selectedValue);
  const formattedSelection = hasValue
    ? formatSelectedDates(selectedValue, undefined, enUS)
    : placeholder;
  const defaultMonth = startOfMonth(selectedValue ?? maxDate ?? today);

  const isClearEnabled = enableClear && !disabled;

  const handleReset = useCallback(() => {
    if (formikContext) {
      formikContext.setFieldValue(props.name, null);
      formikContext.validateField(props.name);
    }

    setSelectedValue(undefined);
  }, [setSelectedValue, formikContext, props.name]);

  const handleSelect = useCallback((value: Date) => {
    if (formikContext) {
      formikContext.setFieldValue(props.name, value);
      formikContext.validateField(props.name);
    }

    setSelectedValue(value);
    close();
  }, [setSelectedValue, formikContext, props.name]);

  const classNames = mergeTailwind(
    {
      'dark:focus:ring-dark-monetr-brand': !props.disabled && !props.error,
      'dark:hover:ring-zinc-400': !props.disabled && !props.error,
      'dark:ring-dark-monetr-border-string': !props.disabled && !props.error,
      'dark:ring-red-500': !props.disabled && !!props.error,
      'ring-gray-300': !props.disabled && !props.error,
      'ring-red-300': !props.disabled && !!props.error,
    },
    {
      'focus:ring-purple-400': !props.error,
      'focus:ring-red-400': props.error,
    },
    {
      'dark:bg-dark-monetr-background': !props.disabled,
      'dark:text-zinc-200': !props.disabled,
      'text-gray-900': !props.disabled,
    },
    { // If there is not a value, the change the text of the button to be 400 for the placeholder.
      'dark:text-gray-400': !hasValue,
    },
    {
      'dark:bg-dark-monetr-background-subtle': props.disabled,
      'dark:ring-dark-monetr-background-emphasis': props.disabled,
      'ring-gray-200': props.disabled,
      'text-gray-500': props.disabled,
    },
    'block',
    'border-0',
    'dark:caret-zinc-50',
    'focus:ring-2',
    'focus:ring-inset',
    'min-h-[38px]',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'text-left',
    'text-sm',
    'w-full',
    'relative',
    className,
  );

  const LabelDecorator = props.labelDecorator || (() => null);
  function Error() {
    if (!props.error) return null;

    return (
      <p className="text-xs font-medium text-red-500 mt-0.5">
        {props.error}
      </p>
    );
  }

  const wrapperClassNames = mergeTailwind({
    // This will make it so the space below the input is the same when there is and isn't an error.
    'pb-[18px]': !props.error,
  }, 'relative', props.className);

  return (
    <Popover
      ref={ ref }
      as="div"
      className={ wrapperClassNames }
      { ...other }
    >
      <MLabel
        label={ props.label }
        disabled={ props.disabled }
        htmlFor={ props.id }
        required={ props.required }
      >
        <LabelDecorator name={ props.name } disabled={ props.disabled } />
      </MLabel>
      <Popover.Button
        disabled={ disabled }
        className={ classNames }
      >
        <TodayOutlined className='text-lg mr-2' />
        <span className="truncate">{formattedSelection}</span>
        { isClearEnabled && selectedValue ? (
          <button
            type="button"
            className={ mergeTailwind(
              'absolute outline-none inset-y-0 right-2 flex items-center transition duration-100 dark:text-dark-monetr-content-subtle',
            ) }
            onClick={ e => {
              e.preventDefault();
              handleReset();
            } }
          >
            <CloseOutlined />
          </button>
        ) : null }
      </Popover.Button>
      <Popover.Panel
        className={ mergeTailwind(
          // common
          'absolute z-50 divide-y overflow-y-auto min-w-min outline-none rounded-lg p-3 border',
          // dark
          'dark:bg-dark-monetr-background dark:border-dark-monetr-border-subtle dark:shadow-2xl',
        ) }
      >
        {({ close }) => (
          <MCalendar<DayPickerSingleProps>
            showOutsideDays={ true }
            mode="single"
            defaultMonth={ defaultMonth }
            selected={ selectedValue }
            onSelect={ (value: Date) => {
              handleSelect(value);
              close();
            } }
            locale={ enUS }
            disabled={ disabledDays }
            enableYearNavigation={ enableYearNavigation }
          />
        )}
      </Popover.Panel>
      <Error />
    </Popover>
  );
});

export default MDatePicker;

export function formatSelectedDates(
  startDate: Date | null,
  endDate: Date | null,
  locale: Locale,
) {
  const localeCode = locale.code;
  if (!startDate && !endDate) {
    return '';
  }

  if (startDate && !endDate) {
    const options: Intl.DateTimeFormatOptions = {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    };
    return startDate.toLocaleDateString(localeCode, options);
  }

  if (startDate && endDate) {
    if (isEqual(startDate, endDate)) {
      const options: Intl.DateTimeFormatOptions = {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      };
      return startDate.toLocaleDateString(localeCode, options);
    }

    if (
      startDate.getMonth() === endDate.getMonth() &&
      startDate.getFullYear() === endDate.getFullYear()
    ) {
      const optionsStartDate: Intl.DateTimeFormatOptions = {
        month: 'short',
        day: 'numeric',
      };
      // eslint-disable-next-line max-len
      return `${startDate.toLocaleDateString(localeCode, optionsStartDate)} - ${endDate.getDate()}, ${endDate.getFullYear()}`;
    }
    const options: Intl.DateTimeFormatOptions = {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    };
    return `${startDate.toLocaleDateString(localeCode, options)} - ${endDate.toLocaleDateString(localeCode, options)}`;
  }
  return '';
};
