import React, { useState } from 'react';
import { ActionMeta, components, OnChangeValue, OptionProps, Theme } from 'react-select';
import { FormatOptionLabelMeta } from 'react-select/base';
import CreatableSelect from 'react-select/creatable';
import { lighten } from '@mui/material';

import clsx from 'clsx';
import { useUpdateTransaction } from 'hooks/transactions';
import Transaction from 'models/Transaction';
import theme from 'theme';

interface Props {
  transaction: Transaction;
}

enum TransactionName {
  Original,
  Custom,
}

interface Option {
  readonly label: string;
  readonly value: TransactionName;
}

const TransactionNameEditor = (props: Props): JSX.Element => {
  const { transaction } = props;
  const [loading, setLoading] = useState<boolean>(false);
  const updateTransaction = useUpdateTransaction();

  async function updateTransactionName(input: string): Promise<void> {
    setLoading(true);
    const updated = new Transaction({
      ...transaction,
      name: input,
    });

    return updateTransaction(updated)
      .catch(error => alert(error?.response?.data?.error || 'Could not save transaction name.'))
      .finally(() => setLoading(false));
  }

  function handleTransactionNameChange(newValue: OnChangeValue<Option, false>, _: ActionMeta<Option>) {
    return updateTransactionName(newValue.label);
  }

  const originalTransactionName: Option = {
    label: transaction.getOriginalName(),
    value: TransactionName.Original,
  };

  const options: Option[] = [
    originalTransactionName,
  ];

  let customTransactionName: Option | null = null;
  if (transaction.name && transaction.name !== transaction.getOriginalName()) {
    customTransactionName = {
      label: transaction.name,
      value: TransactionName.Custom,
    };

    options.push(customTransactionName);
  }

  const createLabelFormat = (inputValue: string): string => {
    return `Rename To: ${ inputValue }`;
  };

  const formatOptionsLabel = (option: Option, meta: FormatOptionLabelMeta<Option>): React.ReactNode => {
    if (meta.context === 'value') {
      return option.label;
    }

    switch (option.value) {
      case TransactionName.Custom:
        return `Custom: ${ option.label }`;
      case TransactionName.Original:
        return `Original: ${ option.label }`;
      default:
        return option.label;
    }
  };

  const value = customTransactionName || originalTransactionName;

  const uiTheme = theme;

  return (
    <CreatableSelect
      theme={ (theme: Theme): Theme => ({
        ...theme,
        colors: {
          ...theme.colors,
          primary: uiTheme.palette.primary.main,
          ...(uiTheme.palette.mode === 'dark' && {
            neutral0: uiTheme.palette.background.default,
            primary25: lighten(uiTheme.palette.background.default, 0.1),
            primary50: lighten(uiTheme.palette.background.default, 0.5),
            neutral80: 'white',
            neutral90: 'white',
          }), // Text
        },
      }) }
      components={ {
        Option: NameSelectionOption,
      } }
      classNamePrefix="transaction-select"
      className="w-full md:basis-1/2 pl-0 pr-0 md:pl-2.5 md:pt-0 md:pb-0 transaction-item-name"
      createOptionPosition="first"
      formatCreateLabel={ createLabelFormat }
      formatOptionLabel={ formatOptionsLabel }
      isClearable={ false }
      isDisabled={ false }
      isLoading={ loading }
      onChange={ handleTransactionNameChange }
      onCreateOption={ updateTransactionName }
      options={ options }
      value={ value }
    />
  );
};

export default TransactionNameEditor;


export interface NameOption {
  readonly label: string;
  readonly value: number | null;
}

export function NameSelectionOption({ children, ...props }: OptionProps<Option>): JSX.Element {
  const optionClassNames = clsx(props.className);
  const labelClassNames = clsx(
    'font-medium',
  );

  return (
    <components.Option { ...props } className={ optionClassNames }>
      <span className={ labelClassNames }>{ props.label }</span>
    </components.Option>
  );
}
