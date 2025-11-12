import React, { useCallback, useMemo, useState } from 'react';
import { isEqual, type Locale, startOfMonth, startOfToday } from 'date-fns';
import { useFormikContext } from 'formik';
import { Calendar as CalendarIcon, X } from 'lucide-react';

import { Calendar } from '@monetr/interface/components/Calendar';
import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator } from '@monetr/interface/components/Label';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import errorTextStyles from './ErrorText.module.scss';
import datePickerStyles from './FormDatePicker.module.scss';
import inputStyles from './FormTextField.module.scss';
import selectStyles from './Select.module.scss';
import typographyStyles from './Typography.module.scss';

export interface FormDatePickerProps extends Omit<React.HTMLAttributes<HTMLButtonElement>, 'value' | 'defaultValue'> {
  value?: Date;
  min?: Date;
  max?: Date;
  children?: React.ReactNode;
  disabled?: boolean;
  enableClear?: boolean;
  enableYearNavigation?: boolean;
  error?: string;
  label?: string;
  labelDecorator?: LabelDecorator;
  name?: string;
  placeholder?: string;
  required?: boolean;
}

export default function FormDatePicker(props: FormDatePickerProps): JSX.Element {
  const { inTimezone } = useTimezone();
  const today = startOfToday({
    in: inTimezone,
  });
  // Load the locale data files so we can format dates with them.
  const { data: locale, isLoading: localeIsLoading } = useLocale();
  const formikContext = useFormikContext();

  const getFormikError = () => {
    if (!formikContext?.touched[props?.name]) {
      return null;
    }

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

  // biome-ignore lint/correctness/useExhaustiveDependencies: Cannot include inTimezone here as it creates a problem.
  React.useEffect(() => {
    setSelectedValue(value ? inTimezone(value) : undefined);
  }, [value]);

  const open = Boolean(anchorEl);

  const disabledDays = useMemo(() => {
    const disabledDays = [];
    if (minDate) {
      disabledDays.push({ before: minDate });
    }
    if (maxDate) {
      disabledDays.push({ after: maxDate });
    }
    return disabledDays;
  }, [minDate, maxDate]);

  const hasValue = Boolean(selectedValue);

  const defaultMonth = startOfMonth(selectedValue ?? maxDate ?? today, {
    in: inTimezone,
  });
  const isClearEnabled = enableClear && !disabled;

  const handleClick = useCallback((event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  }, []);

  const handleClose = useCallback(() => setAnchorEl(null), []);

  const handleReset = useCallback(() => {
    if (formikContext) {
      formikContext.setFieldValue(props.name, null);
      formikContext.setFieldTouched(props.name, true);
      formikContext.validateField(props.name);
    }

    setSelectedValue(undefined);
  }, [formikContext, props.name]);

  const handleSelect = useCallback(
    (value: Date | null) => {
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
    },
    [formikContext, handleClose, props.name, inTimezone],
  );

  const LabelDecorator = props.labelDecorator || (() => null);

  if (localeIsLoading) {
    return (
      <div className={mergeTailwind(errorTextStyles.errorTextPadding, props.className)}>
        <Label disabled={props.disabled} htmlFor={props.id} label={props.label} required={props.required}>
          <LabelDecorator disabled={props.disabled} name={props.name} />
        </Label>
        <div className={mergeTailwind(inputStyles.input, selectStyles.selectLoading)} data-error={props.error}>
          <Skeleton className='w-full h-5 mr-2' />
        </div>
        <ErrorText error={props.error} />
      </div>
    );
  }

  const formattedSelection = hasValue ? formatSelectedDates(selectedValue, undefined, locale) : placeholder;

  return (
    <div
      className={mergeTailwind(errorTextStyles.errorTextPadding, 'relative', className)}
      data-testid={props['data-testid']}
    >
      <Label disabled={props.disabled} htmlFor={props.id} label={props.label} required={props.required}>
        <LabelDecorator disabled={props.disabled} name={props.name} />
      </Label>
      <Popover open={open}>
        <PopoverTrigger asChild>
          <button
            className={mergeTailwind(inputStyles.input, datePickerStyles.datePickerButton)}
            disabled={formikContext?.isSubmitting || disabled}
            onClick={handleClick}
            type='button'
          >
            <CalendarIcon />
            <span className={typographyStyles.truncate}>{formattedSelection}</span>
            {isClearEnabled && selectedValue ? (
              <button
                className={mergeTailwind(
                  'absolute outline-none inset-y-0 right-2 flex items-center transition duration-100 dark:text-dark-monetr-content-subtle',
                )}
                onClick={e => {
                  e.preventDefault();
                  handleReset();
                }}
                type='button'
              >
                <X />
              </button>
            ) : null}
          </button>
        </PopoverTrigger>
        <PopoverContent onPointerDownOutside={handleClose}>
          <Calendar
            className='overflow-y-auto outline-none rounded-lg p-3 bg-dark-monetr-background'
            defaultMonth={defaultMonth}
            disabled={disabledDays}
            enableYearNavigation={enableYearNavigation}
            locale={locale}
            mode='single'
            onSelect={(value: Date) => {
              handleSelect(value);
              handleClose();
            }}
            selected={selectedValue}
            showOutsideDays={true}
          />
        </PopoverContent>
      </Popover>
      <ErrorText error={props.error} />
    </div>
  );
}

export function formatSelectedDates(startDate: Date | null, endDate: Date | null, locale: Locale) {
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

    if (startDate.getMonth() === endDate.getMonth() && startDate.getFullYear() === endDate.getFullYear()) {
      const optionsStartDate: Intl.DateTimeFormatOptions = {
        month: 'short',
        day: 'numeric',
      };
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
}
