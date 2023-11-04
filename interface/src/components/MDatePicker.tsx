/* eslint-disable max-len */

import React, { useCallback, useMemo, useState } from 'react';
import { DayPickerSingleProps } from 'react-day-picker';
import { CloseOutlined, TodayOutlined } from '@mui/icons-material';
import { Popover } from '@mui/material';
import { isEqual, startOfMonth, startOfToday } from 'date-fns';
import enUS from 'date-fns/locale/en-US';
import { useFormikContext } from 'formik';

import MCalendar from './MCalendar';
import MLabel, { MLabelDecorator } from './MLabel';
import { ReactElement } from './types';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

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

export default function MDatePicker(props: MDatePickerProps): JSX.Element {
  const today = startOfToday();
  const formikContext = useFormikContext();

  const getFormikError = () => {
    if (!formikContext?.touched[props?.name]) return null;

    return formikContext?.errors[props?.name];
  };
  props = {
    disabled: formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
    ...props,
  };

  const {
    value = formikContext?.values[props.name],
    min: minDate,
    max: maxDate,
    placeholder = 'Select date',
    enableClear = false,
    disabled,
    className,
    enableYearNavigation = false,
  } = props;

  const [selectedValue, setSelectedValue] = useState<Date | null>(value);
  const [anchorEl, setAnchorEl] = React.useState<HTMLButtonElement | null>(null);
  const open = Boolean(anchorEl);

  const handleClick = useCallback((event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  }, [setAnchorEl]);
  const handleClose = useCallback(() => setAnchorEl(null), [setAnchorEl]);

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
      formikContext.setFieldTouched(props.name, true);
      formikContext.validateField(props.name);
    }

    setSelectedValue(undefined);
  }, [setSelectedValue, formikContext, props.name]);

  const handleSelect = useCallback((value: Date | null) => {
    if (formikContext) {
      formikContext.setFieldValue(props.name, value);
      formikContext.setFieldTouched(props.name, true);
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
  }, 'relative', className);

  return (
    <div className={ wrapperClassNames } data-testid={ props['data-testid'] }>
      <MLabel
        label={ props.label }
        disabled={ props.disabled }
        htmlFor={ props.id }
        required={ props.required }
      >
        <LabelDecorator name={ props.name } disabled={ props.disabled } />
      </MLabel>
      <button
        type='button'
        disabled={ formikContext?.isSubmitting || disabled }
        className={ classNames }
        onClick={ handleClick }
        role='none'
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
      </button>
      <Popover
        open={ open }
        anchorEl={ anchorEl }
        onClose={ handleClose }
        transitionDuration={ 200 }
      >
        <MCalendar<DayPickerSingleProps>
          showOutsideDays={ true }
          mode="single"
          defaultMonth={ defaultMonth }
          selected={ selectedValue }
          onSelect={ (value: Date) => {
            handleSelect(value);
            handleClose();
          } }
          locale={ enUS }
          disabled={ disabledDays }
          enableYearNavigation={ enableYearNavigation }
          className={ mergeTailwind(
            // common
            'z-50 overflow-y-auto outline-none rounded-lg p-3 border',
            // dark
            'dark:bg-dark-monetr-background dark:border-dark-monetr-border-subtle dark:shadow-2xl',
          ) }
        />
      </Popover>
      <Error />
    </div>
  );
}

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
