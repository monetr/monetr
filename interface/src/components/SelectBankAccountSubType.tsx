import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { BankAccountSubType } from '@monetr/interface/models/BankAccount';

interface SelectBankAccountSubTypeProps {
  name: string;
  required?: boolean;
  className?: string;
  disabled?: boolean;
}

const options: Array<SelectOption<BankAccountSubType>> = [
  { label: 'Checking', value: BankAccountSubType.Checking },
  { label: 'Savings', value: BankAccountSubType.Savings },
  { label: 'HSA', value: BankAccountSubType.HSA },
  { label: 'CD', value: BankAccountSubType.CD },
  { label: 'Money Market', value: BankAccountSubType.MoneyMarket },
  { label: 'PayPal', value: BankAccountSubType.PayPal },
  { label: 'Prepaid', value: BankAccountSubType.Prepaid },
  { label: 'Cash Management', value: BankAccountSubType.CashManagement },
  { label: 'EBT', value: BankAccountSubType.EBT },
  { label: 'Credit Card', value: BankAccountSubType.CreditCard },
  { label: 'Auto', value: BankAccountSubType.Auto },
  { label: 'Other', value: BankAccountSubType.Other },
];

export default function SelectBankAccountSubType(props: SelectBankAccountSubTypeProps): React.JSX.Element {
  const formikContext = useFormikContext<Record<string, unknown>>();
  const onChange = useCallback(
    (option: SelectOption<BankAccountSubType>) => {
      formikContext.setFieldValue(props.name, option.value);
    },
    [formikContext, props.name],
  );

  const value = options.find(option => option.value === formikContext.values[props.name]);

  return (
    <Select
      className={props.className}
      disabled={Boolean(props.disabled) || formikContext.isSubmitting}
      isLoading={formikContext.isSubmitting}
      label='Account Sub Type'
      name={props.name}
      onChange={onChange}
      options={options}
      placeholder='Select an account sub type...'
      required={props.required}
      value={value}
    />
  );
}
