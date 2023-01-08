import React from 'react';
import { TextField } from '@mui/material';
import { DatePicker, DatePickerProps } from '@mui/x-date-pickers';
import { useFormikContext } from 'formik';
import moment from 'moment';

type MTDate = moment.Moment;

export interface MDatePickerProps<TInputDate> extends Partial<DatePickerProps<TInputDate, MTDate>> {
  name: string;
  label?: string;
  required?: boolean;
}

export default function MDatePicker<T, TInputDate>(props: MDatePickerProps<TInputDate>): JSX.Element {
  const formik = useFormikContext<T>();
  const minDate = !!props.minDate ? props.minDate : moment().startOf('day').add(1, 'day');
  const touched = formik.touched[props.name];
  const value = formik.values[props.name];
  const error = formik.errors[props.name] as string;
  const hasError = Boolean(touched && value && error?.length > 0);
  const helperText = hasError ? error : '';

  function onChange(value: moment.Moment) {
    formik.setFieldValue(props.name, value.startOf('day'));
  }

  return (
    <DatePicker
      { ...props }
      disabled={ formik.isSubmitting || props.disabled }
      minDate={ minDate }
      onChange={ onChange }
      inputFormat="MM/DD/yyyy"
      value={ value }
      renderInput={ params => (
        <TextField
          label={ props.label }
          fullWidth
          required={ props.required }
          { ...params /* error needs to come after this */ }
          error={ hasError }
          helperText={ helperText }
        />
      ) }
    />
  );
}
