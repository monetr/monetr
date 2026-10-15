import { useMemo } from 'react';
import { useFormikContext } from 'formik';
import { PiggyBank, Plus, Receipt, Wallet } from 'lucide-react';
import { rrulestr } from 'rrule';

import Badge from '@monetr/interface/components/Badge';
import Select, {
  type SelectAction,
  type SelectOption,
  type SelectOptionComponentProps,
  type SelectProps,
} from '@monetr/interface/components/Select';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import type { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import { FREE_TO_USE, FreeToUse, SpendingType } from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import { AmountType } from '@monetr/interface/util/amounts';
import type { FormFieldName, FormValues } from '@monetr/interface/util/typescriptEvils';

import styles from './MSelectSpending.module.scss';

type SpendingOption = Pick<Spending | FreeToUse, 'spendingId' | 'spendingType' | 'currentAmount' | 'name'>;

// Remove the props that we do not want to allow the caller to pass in.
type MSelectSpendingBaseProps = Omit<SelectProps<SpendingOption>, 'options' | 'value' | 'onChange' | 'name'>;

export interface MSelectSpendingProps<T extends FormValues<T>> extends MSelectSpendingBaseProps {
  name?: FormFieldName<T, string | null>;
  // excludeFrom will take the name of another item in the form. The value of that item in the form will be excluded
  // from the list of options presented as part of this select. This is used in the transfer dialog to make sure that
  // both the to and the from selects cannot be the same value.
  excludeFrom?: FormFieldName<T, string | null>;
  // transaction is the one being spent. If it is recurring then the select will suggest the expense its charges are
  // linked to, or offer to create one from it if there isn't one yet.
  transaction?: Transaction;
}

export default function MSelectSpending<T extends FormValues<T>>(props: MSelectSpendingProps<T>): React.JSX.Element {
  const formikContext = useFormikContext<T>();
  const {
    data: spending,
    isLoading: isSpendingsLoading,
    isError: isSpendingsError,
    isSuccess: isSpendingsSuccess,
  } = useSpendings();
  const {
    data: balances,
    isLoading: isBalancesLoading,
    isError: isBalancesError,
    isSuccess: isBalancesSuccess,
  } = useCurrentBalance();
  const { data: recurring } = useRecurringTransaction(props.transaction?.transactionRecurringId ?? null);
  const { data: locale } = useLocaleCurrency();

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

  let options: Array<SelectOption<SpendingOption>> = [
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

  // If the recurring charges are linked to an expense then move that expense to the top as a suggestion. Unless it's
  // already the one picked, then there is nothing to suggest.
  let suggestedSpendingId: ID<Spending> | null = null;
  if (recurring && !recurring.deletedAt && recurring.spendingId !== value) {
    suggestedSpendingId = recurring.spendingId;
  }
  const suggested = options.find(item => suggestedSpendingId && item.value.spendingId === suggestedSpendingId);
  if (props.transaction && recurring && suggested) {
    let description = `Linked to your ${props.transaction.getName()} charges`;
    if (recurring.autoMatched) {
      description = `monetr picked this since your recent ${props.transaction.getName()} charges were spent from it`;
    }

    options = [
      {
        ...suggested,
        group: 'Suggested',
        description,
      },
      ...options.filter(item => item !== suggested),
    ];
  }

  const actions: Array<SelectAction> = [];
  if (props.transaction && recurring?.getCanCreateExpense()) {
    const transaction = props.transaction;
    actions.push({
      group: `This charge repeats ${rrulestr(recurring.ruleset).toText()}`,
      label: `Create Expense From ${transaction.getName()}`,
      description: locale?.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored),
      icon: Plus,
      onSelect: () =>
        showNewExpenseModal({
          recurring,
          transaction,
        }),
    });
  }
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
      actions={actions}
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
        <div className={styles.spendingText}>
          <Typography color='emphasis' ellipsis size='md'>
            {props.label}
          </Typography>
          {props.description && (
            <Typography color='subtle' ellipsis size='sm'>
              {props.description}
            </Typography>
          )}
        </div>
      </div>
      <div className={styles.badges}>
        <Badge size='sm'>{amount}</Badge>
      </div>
    </div>
  );
}
