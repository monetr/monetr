/* eslint-disable max-len */

import React, { useCallback, useMemo, useState } from 'react';
import { tz } from '@date-fns/tz';
import { isEqual, Locale, startOfMonth, startOfToday } from 'date-fns';
import { enUS } from 'date-fns/locale/en-US';
import { useFormikContext } from 'formik';
import { Calendar as CalendarIcon, X } from 'lucide-react';

import MLabel, { MLabelDecorator } from './MLabel';
import { ReactElement } from './types';
import { Button } from '@monetr/interface/components/Button';
import { Calendar } from '@monetr/interface/components/Calendar';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MDatePickerProps extends
  Omit<React.HTMLAttributes<HTMLButtonElement>, 'value' | 'defaultValue'>
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
  const { data: timezone } = useTimezone();
  const inTimezone = useMemo(() => tz(timezone), [timezone]);
  const today = startOfToday({
    in: inTimezone,
  });
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
    enableYearNavigation = true,
  } = props;

  const [selectedValue, setSelectedValue] = useState<Date | null>(value);
  const [anchorEl, setAnchorEl] = React.useState<HTMLButtonElement | null>(null);

  React.useEffect(() => {
    setSelectedValue(value ? inTimezone(value) : undefined);
  }, [value, inTimezone]);

  const open = Boolean(anchorEl);

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

  const defaultMonth = startOfMonth(selectedValue ?? maxDate ?? today, {
    in: inTimezone,
  });
  const isClearEnabled = enableClear && !disabled;

  const handleClick = useCallback((event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  }, [setAnchorEl]);

  const handleClose = useCallback(() => setAnchorEl(null), [setAnchorEl]);

  const handleReset = useCallback(() => {
    if (formikContext) {
      formikContext.setFieldValue(props.name, null);
      formikContext.setFieldTouched(props.name, true);
      formikContext.validateField(props.name);
    }

    setSelectedValue(undefined);
  }, [setSelectedValue, formikContext, props.name]);

  const handleSelect = useCallback((value: Date | null) => {
    // If the value is a selected date then cast the date to the account's timezone.
    if (value) {
      value = inTimezone(value);
    }

    // If we are in a formik form boi then propagate the values upwards.
    if (formikContext) {
      formikContext.setFieldValue(props.name, value);
      formikContext.setFieldTouched(props.name, true);
      formikContext.validateField(props.name);
    }

    // Store the selected value (or lack thereof).
    setSelectedValue(value);
    handleClose();
  }, [formikContext, handleClose, inTimezone, props.name]);

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
    // Normally the button has a ring of 2 for hover, but because we want this to look like a text field this should
    // have a ring of 1.
    'enabled:hover:ring-1',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'text-sm',
    'font-normal',
    'w-full',
    'relative',
    'inline-flex',
    'gap-2',
    'justify-start',
  );

  const LabelDecorator = props.labelDecorator || (() => null);
  function Error() {
    if (!props.error) return null;

    return (
      <p className='text-xs font-medium text-red-500 mt-0.5'>
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
      <Popover open={ open }>
        <PopoverTrigger asChild>
          <Button
            variant='outlined'
            size='select'
            disabled={ formikContext?.isSubmitting || disabled }
            className={ mergeTailwind(classNames) }
            onClick={ handleClick }
          >
            <CalendarIcon />
            <span className='truncate'>{formattedSelection}</span>
            { isClearEnabled && selectedValue ? (
              <button
                type='button'
                className={ mergeTailwind(
                  'absolute outline-none inset-y-0 right-2 flex items-center transition duration-100 dark:text-dark-monetr-content-subtle',
                ) }
                onClick={ e => {
                  e.preventDefault();
                  handleReset();
                } }
              >
                <X />
              </button>
            ) : null }
          </Button>
        </PopoverTrigger>
        <PopoverContent onPointerDownOutside={ handleClose }>
          <Calendar
            showOutsideDays={ true }
            mode='single'
            defaultMonth={ defaultMonth }
            selected={ selectedValue }
            onSelect={ (value: Date) => {
              handleSelect(value);
              handleClose();
            } }
            locale={ enUS }
            disabled={ disabledDays }
            enableYearNavigation={ enableYearNavigation }
            className='overflow-y-auto outline-none rounded-lg p-3 bg-dark-monetr-background'
          />
        </PopoverContent>
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
