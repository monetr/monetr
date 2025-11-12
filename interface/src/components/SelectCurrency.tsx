import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { useInstalledCurrencies } from '@monetr/interface/hooks/useInstalledCurrencies';

interface SelectCurrencyProps {
  name: string;
  required?: boolean;
  className?: string;
  menuPortalTarget?: HTMLElement;
  disabled?: boolean;
}

export default function SelectCurrency(props: SelectCurrencyProps): JSX.Element {
  const formikContext = useFormikContext();
  const { data: currencies, isLoading: currenciesLoading } = useInstalledCurrencies();
  const onChange = useCallback(
    (option: SelectOption<string>) => {
      formikContext.setFieldValue(props.name, option.value);
    },
    [formikContext, props.name],
  );

  if (currenciesLoading) {
    return (
      <Select
        className={props.className}
        disabled
        isLoading
        label='Currency'
        onChange={onChange}
        options={[]}
        placeholder='Select a currency...'
      />
    );
  }

  const options = (currencies ?? []).map(currency => ({ label: currency, value: currency }));
  const value = options.find(option => option.value === formikContext.values[props.name]);

  return (
    <Select
      className={props.className}
      disabled={props.disabled || formikContext.isSubmitting}
      isLoading={currenciesLoading || formikContext.isSubmitting}
      label='Currency'
      name={props.name}
      onChange={onChange}
      options={options}
      placeholder='Select a currency...'
      required={props.required}
      value={value}
    />
  );
}
