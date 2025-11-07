import { useFormikContext } from 'formik';
import { PiggyBank, Receipt } from 'lucide-react';

import Badge from '@monetr/interface/components/Badge';
import Select, {
  type SelectOption,
  type SelectOptionComponentProps,
  type SelectProps,
} from '@monetr/interface/components/Select';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';

import MSpan from './MSpan';

// Remove the props that we do not want to allow the caller to pass in.
type MSelectSpendingBaseProps = Omit<SelectProps<Spending>, 'options' | 'value' | 'onChange'>;

export interface MSelectSpendingProps extends MSelectSpendingBaseProps {
  // excludeFrom will take the name of another item in the form. The value of that item in the form will be excluded
  // from the list of options presented as part of this select. This is used in the transfer dialog to make sure that
  // both the to and the from selects cannot be the same value.
  excludeFrom?: string;
}

const FREE_TO_USE = 'spnd_freeToUse';

export default function MSelectSpending(props: MSelectSpendingProps): JSX.Element {
  const formikContext = useFormikContext();
  const { data: spending, isLoading, isError } = useSpendings();
  const { data: balances } = useCurrentBalance();

  props = {
    label: 'Spent From',
    placeholder: 'Select a spending item...',
    disabled: formikContext?.isSubmitting,
    ...props,
  };

  if (isLoading) {
    return <Select {...props} options={[]} isLoading disabled placeholder='Loading...' onChange={() => {}} />;
  }
  if (isError) {
    return <Select {...props} options={[]} disabled placeholder='Failed to load spending...' onChange={() => {}} />;
  }

  const freeToUse: SelectOption<Spending> = {
    label: 'Free-To-Use',
    value: new Spending({
      spendingId: FREE_TO_USE,
      // It is possible for the "safe" balance to not be present when switching bank accounts. This is a pseudo race
      // condition. Instead we want to gracefully handle the value not being present initially, and print a nicer string
      // until the balance is loaded.
      currentAmount: balances?.free,
    }),
  };

  const items: Array<SelectOption<Spending>> = spending.map(item => ({
    label: item.name,
    value: item,
  }));

  const excludedFrom = formikContext.values[props.excludeFrom];

  const options: Array<SelectOption<Spending>> = [
    freeToUse,
    // Labels will be unique. So we only need 1 | -1
    ...items.sort((a, b) => (a.label.toLowerCase() > b.label.toLowerCase() ? 1 : -1)),
  ].filter(item => {
    // If we are excluding some items and the excluded from has a value from formik.
    // Then make sure our option list omits that item with that value.
    if (props.excludeFrom && excludedFrom) {
      return item.value.spendingId !== excludedFrom;
    }

    // If we are exclluding some items and the excluded item is null(ish) then that means
    // some other select has already picked the safe to spend option. We need to omit that
    // from our result set.
    if (props.excludeFrom && !excludedFrom) {
      return item.value.spendingId !== FREE_TO_USE;
    }

    return true;
  });

  const value: string = formikContext.values[props.name];
  // Determine the current value, if there is not a current value then use null. Null here represents Free to use, which
  // is a non existant spending item that we patch in to represent an unbudgeted transaction.
  const current = options.find(item => item.value.spendingId === (value ?? FREE_TO_USE));

  function onSelect(newValue: SelectOption<Spending>) {
    if (newValue.value.spendingId === FREE_TO_USE) {
      return formikContext.setFieldValue(props.name, null);
    }

    return formikContext.setFieldValue(props.name, newValue.value.spendingId);
  }

  return (
    <Select
      {...props}
      value={current}
      options={options}
      onChange={onSelect}
      optionComponent={SelectSpendingOptionComponent}
    />
  );
}

function SelectSpendingOptionComponent(props: SelectOptionComponentProps<Spending>): React.JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const notLoaded = props.value?.currentAmount === undefined;
  const amount = notLoaded ? 'N/A' : locale.formatAmount(props.value.currentAmount, AmountType.Stored);
  return (
    <div className='flex justify-between'>
      <MSpan size='md' color='emphasis'>
        {props.label}
      </MSpan>
      <div className='flex gap-2'>
        {props.value?.spendingType === SpendingType.Goal && (
          <Badge size='sm' className='dark:bg-dark-monetr-blue  max-h-[24px]'>
            <PiggyBank />
          </Badge>
        )}
        {props.value?.spendingType === SpendingType.Expense && (
          <Badge size='sm' className='dark:bg-dark-monetr-green max-h-[24px]'>
            <Receipt />
          </Badge>
        )}
        <Badge size='sm'>{amount}</Badge>
      </div>
    </div>
  );
}
