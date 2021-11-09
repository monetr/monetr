import Transaction from 'models/Transaction';
import React, { useState } from 'react';
import { useSelector, useStore } from 'react-redux';
import { ActionMeta, OnChangeValue, Theme } from 'react-select';
import { FormatOptionLabelMeta } from 'react-select/base';
import CreatableSelect from 'react-select/creatable';
import updateTransaction from 'shared/transactions/actions/updateTransaction';
import { getTransactionById } from 'shared/transactions/selectors/getTransactionById';

interface PropTypes {
  transactionId: number;
}

enum TransactionName {
  Original,
  Custom,
}

interface Option {
  readonly label: string;
  readonly value: TransactionName;
}

const TransactionNameEditor = (props: PropTypes): JSX.Element => {
  const transaction = useSelector(getTransactionById(props.transactionId));
  const store = useStore();
  const [loading, setLoading] = useState<boolean>();

  const updateTransactionName = (input: string): Promise<void> => {
    setLoading(true);
    const updated = new Transaction({
      ...transaction,
      name: input,
    });

    return updateTransaction(updated)(store.dispatch, store.getState)
      .catch(error => alert(error?.response?.data?.error || 'Could not save transaction name.'))
      .finally(() => setLoading(false));
  };

  const handleTransactionNameChange = (newValue: OnChangeValue<Option, false>, meta: ActionMeta<Option>) => {
    return updateTransactionName(newValue.label)
  };

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
  }

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
  }

  const value = customTransactionName || originalTransactionName;

  return (
    <CreatableSelect
      theme={ (theme: Theme): Theme => ({
        ...theme,
        colors: {
          ...theme.colors,
          primary: '#4E1AA0',
        },
      }) }
      classNamePrefix="transaction-name-select"
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
}

export default TransactionNameEditor;
