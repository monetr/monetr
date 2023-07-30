import React, { useEffect, useState } from 'react';
import { useFormikContext } from 'formik';

import MSelect, { MSelectProps } from './MSelect';

import getRecurrencesForDate from './Recurrence/getRecurrencesForDate';
import Recurrence from './Recurrence/Recurrence';
import { ActionMeta, OnChangeValue } from 'react-select';

export interface MSelectFrequencyProps extends MSelectProps<Recurrence> {
  name: string;
  label: string;
  dateFrom: string;
}

export default function MSelectFrequency(props: MSelectFrequencyProps): JSX.Element {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const formikContext = useFormikContext();

  const date = formikContext.values[props.dateFrom];

  const rules = getRecurrencesForDate(date);

  useEffect(() => {
    if (selectedIndex === rules.length) {
      setSelectedIndex(null);
      formikContext?.setFieldValue(props.name, null);
      return;
    }

    formikContext?.setFieldValue(props.name, rules[selectedIndex]);
  }, [date]);

  const options = rules.map((item, index) => ({
    label: item.name,
    value: index,
  }));

  const value = selectedIndex !== null && selectedIndex >= 0 && selectedIndex < options.length ?
    options[selectedIndex] :
    { label: 'Select a frequency...', value: -1 };

  function onChange(newValue: OnChangeValue<SelectOption, false>, _: ActionMeta<SelectOption>) {
    setSelectedIndex(newValue.value);
    formikContext?.setFieldValue(props.name, rules[newValue.value]);
  }

  return (
    <MSelect
      { ...props }
      onChange={ onChange }
      label={ props.label }
      name={ props.name }
      options={ options }
      isClearable={ false }
      value={ value }
    />
  );
}


interface SelectOption {
  readonly label: string;
  readonly value: number;
}
