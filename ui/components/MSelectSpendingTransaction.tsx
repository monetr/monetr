import React, { Fragment } from 'react';
import Select, { ActionMeta, components, FormatOptionLabelMeta, OnChangeValue, OptionProps, Theme } from 'react-select';

import MBadge from './MBadge';
import MSpan from './MSpan';

import { useCurrentBalance } from 'hooks/balances';
import { useSpendingSink } from 'hooks/spending';
import { useUpdateTransaction } from 'hooks/transactions';
import useTheme from 'hooks/useTheme';
import Spending from 'models/Spending';
import Transaction from 'models/Transaction';
import formatAmount from 'util/formatAmount';
import mergeTailwind from 'util/mergeTailwind';

import './MSelectSpendingTransaction.scss';

export interface MSelectSpendingTransactionProps {
  transaction: Transaction;
}

export default function MSelectSpendingTransaction(props: MSelectSpendingTransactionProps): JSX.Element {
  const { transaction } = props;
  const { result: allSpending } = useSpendingSink();
  const balances = useCurrentBalance();
  const updateTransaction = useUpdateTransaction();
  const theme = useTheme();

  function updateSpentFrom(selection: Spending | null) {
    const spendingId = selection ? selection.spendingId : null;

    // Not strict equal because undefined vs null stuff.
    // If the selected spending is the same as what we have now, do nothing.
    if (spendingId == transaction.spendingId) {
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

  const freeToUse = {
    label: 'Free-To-Use',
    value: null,
    spending: {
      // It is possible for the "safe" balance to not be present when switching bank accounts. This is a pseudo race
      // condition. Instead we want to gracefully handle the value not being present initially, and print a nicer string
      // until the balance is loaded.
      currentAmount: balances?.free,
    },
  };
  const items = allSpending.map(item => ({
    label: item.name,
    value: item.spendingId,
    spending: item,
  }));

  const options = [
    freeToUse,
    // Labels will be unique. So we only need 1 | -1
    ...items.sort((a, b) => a.label.toLowerCase() > b.label.toLowerCase() ? 1 : -1),
  ];

  const selectedItem = !transaction.spendingId ? freeToUse : items.find(item => item.value === transaction.spendingId);

  function formatOptionsLabel(option: SpendingOption, meta: FormatOptionLabelMeta<SpendingOption>): React.ReactNode {
    if (meta.context === 'value') {
      return (
        <Fragment>
          Spent From <span className={ mergeTailwind({
            'font-bold dark:text-dark-monetr-content-emphasis': !!option.value,
          }) }>
            {option.label}
          </span>
        </Fragment>
      );
    }

    return option.label;
  }

  return (
    <div className='hidden md:flex w-1/2 flex-1 items-center'>
      <Select
        theme={ (baseTheme: Theme): Theme => ({
          ...baseTheme,
          borderRadius: 8,
          spacing: {
            controlHeight: 32,
            baseUnit: 3,
            menuGutter: 15,
          },
          colors: {
            ...baseTheme.colors,
            neutral0: theme.tailwind.colors['dark-monetr']['background']['DEFAULT'],
            neutral5: theme.tailwind.colors['dark-monetr']['background']['subtle'],
            neutral10: theme.tailwind.colors['dark-monetr']['background']['emphasis'],
            neutral20: theme.tailwind.colors['dark-monetr']['border']['string'],
            neutral30: theme.tailwind.colors['dark-monetr']['content']['DEFAULT'],
            neutral60: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            neutral70: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            neutral80: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            neutral90: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            primary25: theme.tailwind.colors['dark-monetr']['background']['emphasis'],
            primary50: theme.tailwind.colors['dark-monetr']['brand']['faint'],
            primary: theme.tailwind.colors['dark-monetr']['brand']['DEFAULT'],
          },
        }) }
        menuPortalTarget={ document.body }
        components={ {
          Option: SpendingSelectOption,
        } }
        styles={ {
          placeholder: (base: object) => ({
            ...base,
            fontSize: '0.875rem', // Equivalent to text-sm and leading-6
            lineHeight: '1.5rem',
          }),
          valueContainer: (base: object) => ({
            ...base,
            fontSize: '0.875rem', // Equivalent to text-sm and leading-6
            lineHeight: '1.5rem',
            padding: '0px 5px',
          }),
          option: (base: object) => ({
            ...base,
            color: theme.tailwind.colors['dark-monetr']['content']['emphasized'],
          }),
          menuPortal: (base: object) => ({
            ...base,
            zIndex: 9999,
          }),
        } }
        classNamePrefix='m-select-spending-transaction'
        onChange={ handleSpentFromChange }
        formatOptionLabel={ formatOptionsLabel }
        options={ options }
        value={ selectedItem }
        className='w-full'
      />
    </div>
  );
}

export interface SpendingOption {
  readonly label: string;
  readonly value: number | null;
  readonly spending: Spending | null;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars, unused-imports/no-unused-vars
export function SpendingSelectOption({ children, ...props }: OptionProps<SpendingOption>): JSX.Element {
  // If the current amount is specified then format the amount, if it is not then use N/A.
  const notLoaded = props.data.spending?.currentAmount === undefined;
  const amount = notLoaded ? 'N/A' : formatAmount(props.data.spending.currentAmount);
  return (
    <components.Option { ...props }>
      <div className="flex justify-between">
        <MSpan size='md' color='emphasis'>
          { props.label }
        </MSpan>
        <MBadge size='sm'>
          { amount }
        </MBadge>
      </div>
    </components.Option>
  );
}
