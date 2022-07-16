import React from 'react';
import { components, OptionProps } from 'react-select';
import { Chip } from '@mui/material';

import Spending from 'models/Spending';
import formatAmount from 'util/formatAmount';

export interface SpendingOption {
  readonly label: string;
  readonly value: number | null;
  readonly spending: Spending | null;
}

export function SpendingSelectOption({ children, ...props }: OptionProps<SpendingOption>): JSX.Element {
  // If the current amount is specified then format the amount, if it is not then use N/A.
  const notLoaded = props.data.spending?.currentAmount === undefined;
  const amount = notLoaded ? 'N/A' : formatAmount(props.data.spending.currentAmount);
  return (
    <components.Option { ...props }>
      <div className="w-full flex items-center">
        <span className="font-semibold">{ props.label }</span>
        <Chip
          className="ml-auto font-semibold"
          label={ amount }
          color="secondary"
        />
      </div>
    </components.Option>
  );
}
