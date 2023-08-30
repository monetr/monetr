import React, { useEffect, useState } from 'react';
import { ActionMeta, OnChangeValue } from 'react-select';
import { useFormikContext } from 'formik';

import MSelect, { MSelectProps } from './MSelect';

import getRecurrencesForDate from './Recurrence/getRecurrencesForDate';
import Recurrence from './Recurrence/Recurrence';

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

  // When we initially mount, we want to find the frequency that is currently selected.
  useEffect(() => {
    const currentValue: string = formikContext?.values[props.name];
    const found = rules.findIndex(item => item.ruleString() === currentValue);
    if (found >= 0) {
      // eslint-disable-next-line no-console
      console.log('[MSelectFrequency]', 'found existing recurrence with the specified value');
      setSelectedIndex(found);
      return;
    }

    // eslint-disable-next-line no-console
    console.log('[MSelectFrequency]', 'could not find a recurrence rule:', currentValue);
    // I only want to run this hook when the component mounts.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (selectedIndex === null) return;

    if (selectedIndex >= rules.length) {
      console.log('[MSelectFrequency]', 'date selection has changed and is no longer present in the rules');
      setSelectedIndex(null);
      formikContext?.setFieldValue(props.name, null);
      formikContext?.validateField(props.name);
      return;
    }

    formikContext?.setFieldValue(props.name, rules[selectedIndex].ruleString());
    // I only want to run this hook when the date prop changes. Selected index should not cause this to re-evaluate.
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
    formikContext?.setFieldValue(props.name, rules[newValue.value].ruleString());
  }

  return (
    <MSelect
      { ...props }
      disabled={ formikContext?.isSubmitting }
      error={ formikContext?.errors[props.name] }
      isClearable={ false }
      label={ props.label }
      name={ props.name }
      onChange={ onChange }
      options={ options }
      value={ value }
    />
  );
}


interface SelectOption {
  readonly label: string;
  readonly value: number;
}
