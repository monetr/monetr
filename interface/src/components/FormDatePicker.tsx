import React, { useCallback, useMemo, useState } from 'react';
import { tz } from '@date-fns/tz';
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
  const { data: timezone } = useTimezone();
  const inTimezone = useMemo(() => tz(timezone), [timezone]);
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

  React.useEffect(() => {
    setSelectedValue(value ? inTimezone(value) : undefined);
  }, [value, inTimezone]);

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
    [formikContext, handleClose, inTimezone, props.name],
  );

  const LabelDecorator = props.labelDecorator || (() => null);

  if (localeIsLoading) {
    return (
      <div className={mergeTailwind(errorTextStyles.errorTextPadding, props.className)}>
        <Label label={props.label} disabled={props.disabled} htmlFor={props.id} required={props.required}>
          <LabelDecorator name={props.name} disabled={props.disabled} />
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
      <Label label={props.label} disabled={props.disabled} htmlFor={props.id} required={props.required}>
        <LabelDecorator name={props.name} disabled={props.disabled} />
      </Label>
      <Popover open={open}>
        <PopoverTrigger asChild>
          <button
            type='button'
            disabled={formikContext?.isSubmitting || disabled}
            className={mergeTailwind(inputStyles.input, datePickerStyles.datePickerButton)}
            onClick={handleClick}
          >
            <CalendarIcon />
            <span className={typographyStyles.truncate}>{formattedSelection}</span>
            {isClearEnabled && selectedValue ? (
              <button
                type='button'
                className={mergeTailwind(
                  'absolute outline-none inset-y-0 right-2 flex items-center transition duration-100 dark:text-dark-monetr-content-subtle',
                )}
                onClick={e => {
                  e.preventDefault();
                  handleReset();
                }}
              >
                <X />
              </button>
            ) : null}
          </button>
        </PopoverTrigger>
        <PopoverContent onPointerDownOutside={handleClose}>
          <Calendar
            showOutsideDays={true}
            mode='single'
            defaultMonth={defaultMonth}
            selected={selectedValue}
            onSelect={(value: Date) => {
              handleSelect(value);
              handleClose();
            }}
            locale={locale}
            disabled={disabledDays}
            enableYearNavigation={enableYearNavigation}
            className='overflow-y-auto outline-none rounded-lg p-3 bg-dark-monetr-background'
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
}
