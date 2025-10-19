import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import MSelect from '@monetr/interface/components/MSelect';
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
      console.log('CHANGED', option);
      formikContext.setFieldValue(props.name, option.value);
    },
    [formikContext, props.name],
  );

  if (currenciesLoading) {
    return (
      <Select
        disabled
        label='Currency'
        options={[]}
        isLoading
        placeholder='Select a currency...'
        className={props.className}
        onChange={onChange}
      />
    );
  }

  const options = (currencies ?? []).map(currency => ({ label: currency, value: currency }));
  const value = options.find(option => option.value === formikContext.values[props.name]);

  return (
    <Select
      disabled={props.disabled || formikContext.isSubmitting}
      label='Currency'
      name={props.name}
      onChange={onChange}
      options={options}
      isLoading={currenciesLoading || formikContext.isSubmitting}
      placeholder='Select a currency...'
      required={props.required}
      className={props.className}
      value={value}
    />
  );
}
