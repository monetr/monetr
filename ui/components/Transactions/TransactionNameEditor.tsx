import Transaction from 'data/Transaction';
import React, { Component } from "react";
import { connect } from 'react-redux';
import { ActionMeta, OnChangeValue, Theme } from 'react-select';
import { FormatOptionLabelMeta } from 'react-select/base';
import CreatableSelect from 'react-select/creatable';
import updateTransaction from 'shared/transactions/actions/updateTransaction';
import { getTransactionById } from 'shared/transactions/selectors/getTransactionById';

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  updateTransaction: (transaction: Transaction) => Promise<void>;
}

interface State {
  loading: boolean;
}

enum TransactionName {
  Original,
  Custom,
};

interface Option {
  readonly label: string;
  readonly value: TransactionName;
}

class TransactionNameEditor extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: false,
  };

  handleTransactionNameChange = (newValue: OnChangeValue<Option, false>, meta: ActionMeta<Option>) => {
    return this.updateTransactionName(newValue.label)
  };

  updateTransactionName = (input: string): Promise<void> => {
    this.setState({
      loading: true,
    });
    const { transaction, updateTransaction } = this.props;
    const updated = new Transaction({
      ...transaction,
      name: input,
    });

    return updateTransaction(updated)
      .catch(error => alert(error?.response?.data?.error || 'Could not save transaction name.'))
      .finally(() => this.setState({
        loading: false,
      }));
  };

  render() {
    const { transaction } = this.props;

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
        isLoading={ this.state.loading }
        onChange={ this.handleTransactionNameChange }
        onCreateOption={ this.updateTransactionName }
        options={ options }
        value={ value }
      />
    );
  }
}

export default connect(
  (state, props: PropTypes) => ({
    transaction: getTransactionById(props.transactionId)(state)
  }),
  {
    updateTransaction,
  }
)(TransactionNameEditor);
