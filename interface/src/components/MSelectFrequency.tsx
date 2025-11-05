import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFormikContext } from 'formik';

import Select, { type SelectOption, type SelectProps } from '@monetr/interface/components/Select';
import useTimezone from '@monetr/interface/hooks/useTimezone';

import getRecurrencesForDate from './Recurrence/getRecurrencesForDate';

export interface MSelectFrequencyProps extends Omit<SelectProps<string>, 'onChange' | 'options'> {
  name: string;
  label: string;
  dateFrom: string;
}

export default function MSelectFrequency(props: MSelectFrequencyProps): JSX.Element {
  const { data: timezone } = useTimezone();
  const [selectedSignature, setSelectedSignature] = useState<string | null>(null);
  const formikContext = useFormikContext();

  const date = formikContext.values[props.dateFrom];

  const rules = useMemo(() => getRecurrencesForDate(date, timezone), [date, timezone]);
  const options = useMemo(
    () =>
      rules.map(item => ({
        label: item.name,
        value: item.signature(),
      })),
    [rules],
  );

  const value = useMemo(() => {
    return options.find(option => option.value === selectedSignature) ?? null;
  }, [options, selectedSignature]);

  // Every time the date input changes we need to rebuild the list of recurrences. When this happens we should also try
  // to find a recurrence that matches our current rule. This happens when we have a rule like the 15,-1 and the current
  // date is the 15th, and the user changes it to -1. The rule is still valid even though the date has changed. But in
  // any other scenario where the rule is no longer valid. We want to remove the selection and make sure they provide a
  // new frequency.
  // biome-ignore lint/correctness/useExhaustiveDependencies: I want to only re-run this hook when the date prop changes
  useEffect(() => {
    const currentValue: string = formikContext.values[props.name];
    const found = currentValue ? rules.find(rule => rule.equalSignature(currentValue)) : null;
    if (found) {
      setSelectedSignature(found.signature());
      return;
    }

    setSelectedSignature(null);
    formikContext.setFieldValue(props.name, null);
    formikContext.validateField(props.name);
    // I only want to run this hook when the date prop changes. Selected index should not cause this to re-evaluate.
  }, [date]);

  const onChange = useCallback(
    async (newValue: SelectOption<string>) => {
      setSelectedSignature(newValue.value);
      await formikContext
        .setFieldValue(props.name, rules.find(rule => rule.signature() === newValue.value).ruleString())
        // After the value has been set, we need to set the field as touched and trigger a revalidation of that specific
        // field.
        .then(async () => void (await formikContext.setFieldTouched(props.name, true)))
        .then(async () => void (await formikContext.validateField(props.name)));
    },
    [props.name, formikContext, rules],
  );

  return (
    <Select<string>
      {...props}
      placeholder='Select a frequency...'
      disabled={props.disabled || formikContext.isSubmitting}
      error={formikContext.errors[props.name]}
      label={props.label}
      name={props.name}
      onChange={onChange}
      options={options}
      value={value}
    />
  );
}
