import React, { Fragment } from 'react';
import Select, { ActionMeta, OnChangeValue, Theme } from 'react-select';
import { FormatOptionLabelMeta } from 'react-select/base';
import classnames from 'classnames';

import { SpendingOption, SpendingSelectOption } from 'components/Transactions/components/SpendingSelectOption';
import { useCurrentBalance } from 'hooks/balances';
import { useSpendingSink } from 'hooks/spending';
import { useUpdateTransaction } from 'hooks/transactions';
import Spending from 'models/Spending';
import Transaction from 'models/Transaction';

interface Props {
  transaction: Transaction;
}

function TransactionSpentFromSelection(props: Props): JSX.Element {
  const { transaction } = props;
  const { result: allSpending } = useSpendingSink();
  const balances = useCurrentBalance();
  const updateTransaction = useUpdateTransaction();

  if (transaction.getIsAddition()) {
    return (
      <span
        className="flex items-center w-full md:basis-1/2 pl-3 pr-0 mt-2.5 md:pl-5 md:mt-0 md:mb-0 opacity-50"
        style={ { height: '38px' } }
      >
        Deposit
      </span>
    );
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

  function handleSpentFromChange(newValue: OnChangeValue<SpendingOption, false>, _: ActionMeta<SpendingOption>) {
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
  };
  const items: Map<number, SpendingOption> = new Map(Array.from(allSpending.values())
    .map(item => [item.spendingId, ({
      label: item.name,
      value: item.spendingId,
      spending: item,
    })]));

  const options = [
    safeToSpend,
    // Labels will be unique. So we only need 1 | -1
    ...(Array.from(items.values()).sort((a, b) => a.label.toLowerCase() > b.label.toLowerCase() ? 1 : -1)),
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
      );
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
      classNamePrefix="transaction-select"
      className="w-full md:basis-1/2 pl-0 pr-1 pt-2.5 md:pl-2.5 md:pt-0 md:pb-0"
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

export default React.memo(TransactionSpentFromSelection);
