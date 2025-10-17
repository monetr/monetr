import React, { useCallback } from 'react';
import type { OnChangeValue } from 'react-select';
import { useFormikContext } from 'formik';

import MSelect from '@monetr/interface/components/MSelect';
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
    (option: OnChangeValue<{ label: string; value: string }, false>) => {
      formikContext.setFieldValue(props.name, option.value);
    },
    [formikContext, props.name],
  );

  if (currenciesLoading) {
    return (
      <MSelect
        label='Currency'
        disabled
        isLoading
        placeholder='Select a currency...'
        required={props.required}
        className={props.className}
        menuPortalTarget={props.menuPortalTarget}
      />
    );
  }

  const options = (currencies ?? []).map(currency => ({ label: currency, value: currency }));
  const value = options.find(option => option.value === formikContext.values[props.name]);

  return (
    <MSelect
      disabled={props.disabled || formikContext.isSubmitting}
      label='Currency'
      name={props.name}
      onChange={onChange}
      options={options}
      isLoading={currenciesLoading || formikContext.isSubmitting}
      placeholder='Select a currency...'
      required={props.required}
      className={props.className}
      menuPortalTarget={props.menuPortalTarget}
      value={value}
    />
  );
}
