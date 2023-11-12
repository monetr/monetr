import React from 'react';
import { components, OptionProps } from 'react-select';
import { PriceCheckOutlined, SavingsOutlined } from '@mui/icons-material';
import { useFormikContext } from 'formik';

import MBadge from './MBadge';
import MSelect, { MSelectProps } from './MSelect';
import MSpan from './MSpan';
import { useCurrentBalance } from '@monetr/interface/hooks/balances';
import { useSpendingSink } from '@monetr/interface/hooks/spending';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import { formatAmount } from '@monetr/interface/util/amounts';

// Remove the props that we do not want to allow the caller to pass in.
type MSelecteSpendingBaseProps = Omit<
  MSelectProps<SpendingOption>,
  'options' | 'current' | 'value' | 'onChange' | 'components'
>;

export interface MSelectSpendingProps extends MSelecteSpendingBaseProps {
  // excludeFrom will take the name of another item in the form. The value of that item in the form will be excluded
  // from the list of options presented as part of this select. This is used in the transfer dialog to make sure that
  // both the to and the from selects cannot be the same value.
  excludeFrom?: string;
}

export default function MSelectSpending(props: MSelectSpendingProps): JSX.Element {
  const formikContext = useFormikContext();
  const { result: spending, isLoading, isError } = useSpendingSink();
  const balances = useCurrentBalance();

  props = {
    label: 'Spent From',
    placeholder: 'Select a spending item...',
    disabled: formikContext?.isSubmitting,
    ...props,
  };

  if (isLoading) {
    return <MSelect
      { ...props }
      isLoading
      disabled
      placeholder='Loading...'
    />;
  }
  if (isError) {
    return <MSelect
      { ...props }
      isLoading
      disabled
      placeholder='Failed to load spending...'
    />;
  }

  const freeToUse: SpendingOption = {
    label: 'Free-To-Use',
    value: -1,
    spending: new Spending({
      spendingId: -1,
      // It is possible for the "safe" balance to not be present when switching bank accounts. This is a pseudo race
      // condition. Instead we want to gracefully handle the value not being present initially, and print a nicer string
      // until the balance is loaded.
      currentAmount: balances?.free,
    }),
  };

  const items: Array<SpendingOption> = spending.map(item => ({
    label: item.name,
    value: item.spendingId,
    spending: item,
  }));

  const excludedFrom = formikContext.values[props.excludeFrom];

  const options: Array<SpendingOption> = [
    freeToUse,
    // Labels will be unique. So we only need 1 | -1
    ...items
      .sort((a, b) => a.label.toLowerCase() > b.label.toLowerCase() ? 1 : -1),
  ]
    .filter(item => {
      // If we are excluding some items and the excluded from has a value from formik.
      // Then make sure our option list omits that item with that value.
      if (props.excludeFrom && excludedFrom) {
        return item.value !== excludedFrom;
      }

      // If we are exclluding some items and the excluded item is null(ish) then that means
      // some other select has already picked the safe to spend option. We need to omit that
      // from our result set.
      if (props.excludeFrom && !excludedFrom) {
        return item.value !== -1;
      }

      return true;
    });

  const value = formikContext.values[props.name];
  const current = options.find(item => item.value === (value ?? -1));

  function onSelect(newValue: { label: string, value: number }) {
    if (newValue.value === -1) {
      return formikContext.setFieldValue(props.name, null);
    }

    return formikContext.setFieldValue(props.name, newValue.value);
  }

  return (
    <MSelect
      { ...props }
      options={ options }
      value={ current }
      onChange={ onSelect }
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
  const notLoaded = props.data.spending?.currentAmount === undefined;
  const amount = notLoaded ? 'N/A' : formatAmount(props.data.spending.currentAmount);
  return (
    <components.Option { ...props }>
      <div className='flex justify-between'>
        <MSpan size='md' color='emphasis'>
          { props.label }
        </MSpan>
        <div className='flex gap-2'>
          { props.data.spending?.spendingType === SpendingType.Goal &&
            <MBadge size='sm' className='dark:bg-dark-monetr-blue  max-h-[24px]'>
              <SavingsOutlined />
            </MBadge>
          }
          { props.data.spending?.spendingType === SpendingType.Expense &&
            <MBadge size='sm' className='dark:bg-dark-monetr-green max-h-[24px]'>
              <PriceCheckOutlined />
            </MBadge>
          }
          <MBadge size='sm'>
            {amount}
          </MBadge>
        </div>
      </div>
    </components.Option>
  );
}
