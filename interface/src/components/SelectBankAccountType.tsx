import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { BankAccountType } from '@monetr/interface/models/BankAccount';

interface SelectBankAccountTypeProps {
  name: string;
  required?: boolean;
  className?: string;
  disabled?: boolean;
}

const options: Array<SelectOption<BankAccountType>> = [
  { label: 'Depository', value: BankAccountType.Depository },
  { label: 'Credit', value: BankAccountType.Credit },
  { label: 'Loan', value: BankAccountType.Loan },
  { label: 'Investment', value: BankAccountType.Investment },
  { label: 'Other', value: BankAccountType.Other },
];

export default function SelectBankAccountType(props: SelectBankAccountTypeProps): React.JSX.Element {
  const formikContext = useFormikContext<Record<string, unknown>>();
  const onChange = useCallback(
    (option: SelectOption<BankAccountType>) => {
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
      label='Account Type'
      name={props.name}
      onChange={onChange}
      options={options}
      placeholder='Select an account type...'
      required={props.required}
      value={value}
    />
  );
}
