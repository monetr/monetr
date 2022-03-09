import classnames from 'classnames';
import { SpendingOption, SpendingSelectOption } from 'components/Transactions/components/SpendingSelectOption';
import Spending from 'models/Spending';
import Transaction from 'models/Transaction';
import React, { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { FormatOptionLabelMeta } from 'react-select/base';
import { getBalance } from 'shared/balances/selectors/getBalance';
import { getSpending } from 'shared/spending/selectors/getSpending';
import useUpdateTransaction from 'shared/transactions/actions/updateTransaction';
import { getTransactionById } from 'shared/transactions/selectors/getTransactionById';
import Select, { ActionMeta, OnChangeValue, Theme } from 'react-select';
import { Map } from 'immutable';

import './styles/TransactionSpentFromSelection.scss';

interface Props {
  transactionId: number;
}

export default function TransactionSpentFromSelection(props: Props): JSX.Element {
  const transaction = useSelector(getTransactionById(props.transactionId));
  const allSpending = useSelector(getSpending);
  const balances = useSelector(getBalance);
  const updateTransaction = useUpdateTransaction();

  if (transaction.getIsAddition()) {
    return null;
  }

  function updateSpentFrom(selection: Spending | null) {
    const spendingId = selection ? selection.spendingId : null;

    if (spendingId === transaction.spendingId) {
      return Promise.resolve();
    }

    const updatedTransaction = new Transaction({
      ...transaction,
      spendingId: spendingId,
    });

    return updateTransaction(updatedTransaction);
  }

  function handleSpentFromChange(newValue: OnChangeValue<SpendingOption, false>, meta: ActionMeta<SpendingOption>) {
    return updateSpentFrom(newValue.spending);
  }

  const safeToSpend = {
    label: 'Safe-To-Spend',
    value: null,
    spending: {
      // It is possible for the "safe" balance to not be present when switching bank accounts. This is a pseudo race
      // condition. Instead we want to gracefully handle the value not being present initially, and print a nicer string
      // until the balance is loaded.
      currentAmount: balances?.safe,
    },
  }
  const items: Map<number, SpendingOption> = allSpending
    .sortBy(item => item.name.toLowerCase()) // Sort without case sensitivity.
    .map(item => ({
      label: item.name,
      value: item.spendingId,
      spending: item,
    }));

  const options = [
    safeToSpend,
    ...items.valueSeq().toArray(),
  ];

  const selectedItem = !transaction.spendingId ? safeToSpend : items.get(transaction.spendingId);

  function formatOptionsLabel(option: SpendingOption, meta: FormatOptionLabelMeta<SpendingOption>): React.ReactNode {
    if (meta.context === 'value') {
      return (
        <Fragment>
          Spent From <span className={ classnames({ 'font-bold': !!option.value }) }>
            { option.label }
          </span>
        </Fragment>
      )
    }

    return option.label;
  }

  return (
    <Select
      theme={ (theme: Theme): Theme => ({
        ...theme,
        colors: {
          ...theme.colors,
          primary: '#4E1AA0',
        },
      }) }
      components={ {
        Option: SpendingSelectOption,
      } }
      classNamePrefix="transaction-spending-select"
      isClearable={ false }
      isDisabled={ false }
      isLoading={ false }
      onChange={ handleSpentFromChange }
      formatOptionLabel={ formatOptionsLabel }
      options={ options }
      value={ selectedItem }
    />
  );
}
