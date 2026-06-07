import { useMemo } from 'react';
import { useFormikContext } from 'formik';
import { PiggyBank, Receipt, Wallet } from 'lucide-react';

import Badge from '@monetr/interface/components/Badge';
import Select, {
  type SelectOption,
  type SelectOptionComponentProps,
  type SelectProps,
} from '@monetr/interface/components/Select';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import { FREE_TO_USE, FreeToUse, SpendingType } from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './MSelectSpending.module.scss';

type SpendingOption = Pick<Spending | FreeToUse, 'spendingId' | 'spendingType' | 'currentAmount' | 'name'>;

// Remove the props that we do not want to allow the caller to pass in.
type MSelectSpendingBaseProps = Omit<SelectProps<SpendingOption>, 'options' | 'value' | 'onChange'>;

export interface MSelectSpendingProps extends MSelectSpendingBaseProps {
  // excludeFrom will take the name of another item in the form. The value of that item in the form will be excluded
  // from the list of options presented as part of this select. This is used in the transfer dialog to make sure that
  // both the to and the from selects cannot be the same value.
  excludeFrom?: string;
  bankAccountId: ID<BankAccount>;
}

export default function MSelectSpending(props: MSelectSpendingProps): React.JSX.Element {
  const formikContext = useFormikContext<Record<string, any>>();
  const {
    data: spending,
    isLoading: isSpendingsLoading,
    isError: isSpendingsError,
    isSuccess: isSpendingsSuccess,
  } = useSpendings(props.bankAccountId);
  const {
    data: balances,
    isLoading: isBalancesLoading,
    isError: isBalancesError,
    isSuccess: isBalancesSuccess,
  } = useCurrentBalance(props.bankAccountId);

  props = {
    label: 'Spent From',
    placeholder: 'Select a spending item...',
    disabled: formikContext?.isSubmitting,
    ...props,
  };

  const items: Array<SelectOption<SpendingOption>> = useMemo(() => {
    return (spending ?? []).map(item => ({
      label: item.name,
      value: item,
    }));
  }, [spending]);

  if (isSpendingsLoading || isBalancesLoading) {
    return <Select {...props} disabled isLoading onChange={() => {}} options={[]} placeholder='Loading...' />;
  }
  if (isSpendingsError || isBalancesError) {
    return <Select {...props} disabled onChange={() => {}} options={[]} placeholder='Failed to load spending...' />;
  }
  if (!isSpendingsSuccess || !isBalancesSuccess) {
    return <Select {...props} disabled onChange={() => {}} options={[]} placeholder='Failed to load spending...' />;
  }

  const freeToUse: SelectOption<SpendingOption> = {
    label: 'Free-To-Use',
    value: new FreeToUse(balances),
  };

  const excludedFrom = props.excludeFrom ? formikContext.values[props.excludeFrom] : undefined;

  const options: Array<SelectOption<SpendingOption>> = [
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

  const value: string | undefined = props.name ? formikContext.values[props.name] : undefined;
  // Determine the current value, if there is not a current value then use null. Null here represents Free to use, which
  // is a non existant spending item that we patch in to represent an unbudgeted transaction.
  const current = options.find(item => item.value.spendingId === (value ?? FREE_TO_USE));

  function onSelect(newValue: SelectOption<SpendingOption>) {
    if (!props.name) {
      return;
    }

    if (newValue.value.spendingId === FREE_TO_USE) {
      return formikContext.setFieldValue(props.name, null);
    }

    return formikContext.setFieldValue(props.name, newValue.value.spendingId);
  }

  return (
    <Select
      {...props}
      onChange={onSelect}
      optionComponent={SelectSpendingOptionComponent}
      options={options}
      value={current}
    />
  );
}

export function SelectSpendingOptionComponent(props: SelectOptionComponentProps<SpendingOption>): React.JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const notLoaded = props.value?.currentAmount === undefined;
  const amount = notLoaded || !locale ? 'N/A' : locale.formatAmount(props.value.currentAmount, AmountType.Stored);
  return (
    <div className={styles.optionRow}>
      <div className={styles.spendingName}>
        {props.value?.spendingType === SpendingType.FreeToUse && (
          <Badge className={styles.iconBadge} size='sm' variant='brand'>
            <Wallet />
          </Badge>
        )}
        {props.value?.spendingType === SpendingType.Goal && (
          <Badge className={styles.iconBadge} size='sm' variant='info'>
            <PiggyBank />
          </Badge>
        )}
        {props.value?.spendingType === SpendingType.Expense && (
          <Badge className={styles.iconBadge} size='sm' variant='positive'>
            <Receipt />
          </Badge>
        )}
        <Typography color='emphasis' ellipsis size='md'>
          {props.label}
        </Typography>
      </div>
      <div className={styles.badges}>
        <Badge size='sm'>{amount}</Badge>
      </div>
    </div>
  );
}
