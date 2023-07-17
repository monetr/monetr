import React from 'react';
import { components, OptionProps } from 'react-select';

import MSelect from './MSelect';
import MSpan from './MSpan';

import { useCurrentBalance } from 'hooks/balances';
import { useSpendingSink } from 'hooks/spending';
import Spending from 'models/Spending';

export interface MSelectSpendingProps {
  className?: string;
}

export default function MSelectSpending(props: MSelectSpendingProps): JSX.Element {
  const { result: spending, isLoading, isError } = useSpendingSink();
  const balances = useCurrentBalance();

  if (isLoading) {
    return <MSelect
      className={ props?.className }
      label="Spent From"
      isLoading
      disabled
      placeholder='Loading...'
    />;
  }
  if (isError) {
    return <MSelect
      className={ props?.className }
      label="Spent From"
      isLoading
      disabled
      placeholder='Failed to load spending...'
    />;
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

  const items: Map<number, SpendingOption> = new Map(spending
    .map(item => [item.spendingId, ({
      label: item.name,
      value: item.spendingId,
      spending: item,
    })]));

  const options = [
    freeToUse,
    // Labels will be unique. So we only need 1 | -1
    ...(Array.from(items.values()).sort((a, b) => a.label.toLowerCase() > b.label.toLowerCase() ? 1 : -1)),
  ];

  return (
    <MSelect
      className={ props?.className }
      label="Spent From"
      placeholder='Select a spending item...'
      options={ options }
      name="spendingId"
      components={ {
        Option: MSelectSpendingOption,
      } }
      isClearable={ false }
    />
  );
}

interface SpendingOption {
  readonly label: string;
  readonly value: number | null;
  readonly spending: Spending | null;
}

function MSelectSpendingOption({ children: _, ...props }: OptionProps<SpendingOption>): JSX.Element {
  // const notLoaded = props.data.spending?.currentAmount === undefined;
  // const amount = notLoaded ? 'N/A' : formatAmount(props.data.spending.currentAmount);
  return (
    <components.Option { ...props }>
      <div>
        <MSpan>
          { props.label }
        </MSpan>
      </div>
    </components.Option>
  );
}
