import { useCallback, useEffect, useMemo, useState } from 'react';
import { tz } from '@date-fns/tz';
import { isSameDay } from 'date-fns';
import { useFormikContext } from 'formik';
import { RRule, rrulestr } from 'rrule';

import Select, { type SelectOption, type SelectProps } from '@monetr/interface/components/Select';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import Recurrence from '@monetr/interface/models/Recurrence';
import getRecurrencesForDate from '@monetr/interface/util/getRecurrencesForDate';
import parseDate from '@monetr/interface/util/parseDate';
import type { FormFieldName, FormValues } from '@monetr/interface/util/typescriptEvils';

const customSignature = 'custom';

export interface SelectFrequencyProps<T extends FormValues<T>>
  extends Omit<SelectProps<string>, 'onChange' | 'options' | 'name'> {
  name: FormFieldName<T, string | null>;
  label: string;
  dateFrom: FormFieldName<T, Date | null>;
}

export default function SelectFrequency<T extends FormValues<T>>(props: SelectFrequencyProps<T>): React.JSX.Element {
  const { timezone } = useTimezone();
  const [selectedSignature, setSelectedSignature] = useState<string | null>(null);
  const formikContext = useFormikContext<T>();

  const date = useMemo(() => formikContext.values[props.dateFrom], [props.dateFrom, formikContext.values]);

  const rules = useMemo(() => getRecurrencesForDate(date, timezone), [date, timezone]);

  // The rule the form was loaded with might not match any of the rules we generate for its date, like a rule created by
  // an older version of the UI or one the server generated. Rather than throwing it away we show it as a read only
  // custom option. It is only offered while the date is still the date the form was loaded with.
  const initialRuleString: string | null = formikContext.initialValues[props.name] ?? null;
  const initialDate = formikContext.initialValues[props.dateFrom];
  const customRule = useMemo(() => {
    if (!initialRuleString) {
      return null;
    }
    const current = parseDate(date);
    const initial = parseDate(initialDate);
    if (!current || !initial || !isSameDay(current, initial, { in: tz(timezone) })) {
      return null;
    }
    if (rules.some(rule => rule.equalSignature(initialRuleString))) {
      return null;
    }
    try {
      const rule = rrulestr(initialRuleString);
      if (!(rule instanceof RRule)) {
        return null;
      }
      return new Recurrence({
        name: rule.isFullyConvertibleToText() ? `Custom (${rule.toText()})` : 'Custom schedule',
        rule,
      });
    } catch {
      return null;
    }
  }, [initialRuleString, initialDate, date, timezone, rules]);

  const options = useMemo(
    () => [
      ...(customRule ? [{ label: customRule.name, value: customSignature }] : []),
      ...rules.map(item => ({
        label: item.name,
        value: item.signature(),
      })),
    ],
    [customRule, rules],
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
      formikContext.setFieldValue(props.name, found.ruleString());
      formikContext.setFieldTouched(props.name, true);
      formikContext.validateField(props.name);
      return;
    }

    if (customRule) {
      setSelectedSignature(customSignature);
      formikContext.setFieldValue(props.name, initialRuleString);
      formikContext.validateField(props.name);
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
      // The custom rule is written back exactly as it was stored rather than regenerated from the parsed rule.
      const ruleString =
        newValue.value === customSignature
          ? initialRuleString
          : rules.find(rule => rule.signature() === newValue.value)?.ruleString();
      await formikContext
        .setFieldValue(props.name, ruleString)
        // After the value has been set, we need to set the field as touched and trigger a revalidation of that specific
        // field.
        .then(async () => void (await formikContext.setFieldTouched(props.name, true)))
        .then(async () => void (await formikContext.validateField(props.name)));
    },
    [props.name, formikContext, rules, initialRuleString],
  );

  return (
    <Select<string>
      {...props}
      disabled={Boolean(props.disabled) || formikContext.isSubmitting}
      error={formikContext.errors[props.name] as string | undefined}
      label={props.label}
      name={props.name}
      onChange={onChange}
      options={options}
      placeholder={options.length > 0 ? 'Select a frequency...' : 'Select a date first...'}
      value={value ?? undefined}
    />
  );
}
