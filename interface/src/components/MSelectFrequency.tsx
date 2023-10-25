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

  // Every time the date input changes we need to rebuild the list of recurrences. When this happens we should also try
  // to find a recurrence that matches our current rule. This happens when we have a rule like the 15,-1 and the current
  // date is the 15th, and the user changes it to -1. The rule is still valid even though the date has changed. But in
  // any other scenario where the rule is no longer valid. We want to remove the selection and make sure they provide a
  // new frequency.
  useEffect(() => {
    // eslint-disable-next-line no-console
    console.debug('[MSelectFrequency]', 'date selection has changed and is no longer present in the rules');
    const currentValue: string = formikContext?.values[props.name];
    const found = currentValue ? rules.findIndex(item => item.equalRule(currentValue)) : -1;
    if (found >= 0) {
      // eslint-disable-next-line no-console
      console.debug('[MSelectFrequency]', 'found existing recurrence with the specified value');
      setSelectedIndex(found);
      return;
    }

    setSelectedIndex(null);
    formikContext?.setFieldValue(props.name, null);
    formikContext?.validateField(props.name);
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
    formikContext?.setFieldTouched(props.name, true);
    formikContext?.validateField(props.name);
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
