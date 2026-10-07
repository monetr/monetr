import { useState } from 'react';
import { Plus, Repeat, Sparkles } from 'lucide-react';
import { Link } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import { SelectSpendingOptionComponent } from '@monetr/interface/components/MSelectSpending';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { Switch } from '@monetr/interface/components/Switch';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchTransactionRecurring } from '@monetr/interface/hooks/usePatchTransactionRecurring';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import type { ID } from '@monetr/interface/models/ID';
import { FREE_TO_USE, FreeToUse, type default as Spending, SpendingType } from '@monetr/interface/models/Spending';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { AmountType } from '@monetr/interface/util/amounts';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './RecurringSpendingCard.module.scss';

type SpendingOption = Pick<Spending | FreeToUse, 'spendingId' | 'spendingType' | 'currentAmount' | 'name'>;

export interface RecurringSpendingCardProps {
  recurring: TransactionRecurring;
  name: string;
}

export default function RecurringSpendingCard(props: RecurringSpendingCardProps): React.JSX.Element {
  const { data: spending, isLoading: spendingIsLoading } = useSpendings();
  const { data: balances, isLoading: balancesIsLoading } = useCurrentBalance();
  const { data: locale } = useLocaleCurrency();
  const patchTransactionRecurring = usePatchTransactionRecurring();
  const { enqueueSnackbar } = useSnackbar();
  const [saving, setSaving] = useState(false);

  // Only expenses can be linked for now, goals will get their own thing later
  let options: Array<SelectOption<SpendingOption>> = (spending ?? [])
    .filter(item => item.spendingType === SpendingType.Expense)
    .sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
    .map(item => ({
      label: item.name,
      value: item,
    }));
  const hasExpenses = options.length > 0;

  // Free-To-Use is how you say its not budgeted with anything, same as everywhere else spending gets picked
  if (balances) {
    options = [
      {
        label: 'Free-To-Use',
        value: new FreeToUse(balances),
      },
      ...options,
    ];
  }

  const value = options.find(item => item.value.spendingId === (props.recurring.spendingId ?? FREE_TO_USE));

  let placeholder = 'Select an expense...';
  if (!hasExpenses) {
    placeholder = 'No expenses exist...';
  }

  async function linkSpending(newValue: SelectOption<SpendingOption>) {
    // Picking Free-To-Use means clearing the link, so send null to the server
    let spendingId: ID<Spending> | null = newValue.value.spendingId;
    if (spendingId === FREE_TO_USE) {
      spendingId = null;
    }

    if (spendingId === props.recurring.spendingId) {
      return;
    }

    setSaving(true);
    return await patchTransactionRecurring({
      transactionRecurringId: props.recurring.transactionRecurringId,
      bankAccountId: props.recurring.bankAccountId,
      spendingId,
    })
      .then(
        () =>
          void enqueueSnackbar('Updated recurring transaction successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error?.response?.data?.error || 'Failed to update recurring transaction', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => setSaving(false));
  }

  return (
    <section className={styles.root}>
      <div className={styles.header}>
        <Typography component='h3' size='lg' weight='semibold'>
          Automatically Spend
        </Typography>
        <Typography color='subtle' size='sm'>
          Pick where new {props.name || 'recurring'} charges come out of when they show up.
        </Typography>
      </div>
      <div className={styles.card} data-active={String(Boolean(props.recurring.spendingId))}>
        <div className={styles.cardHeader}>
          <div className={styles.cardText}>
            <span className={styles.chip}>
              <Repeat />
              This Schedule
            </span>
            <span className={styles.title}>Spend Charges On This Schedule From</span>
            <span className={styles.description}>
              Only the {locale?.formatAmount(Math.abs(props.recurring.lastAmount), AmountType.Stored)} charges monetr
              matches to this schedule.
            </span>
          </div>
          {/* Automatically spending isn't a thing yet, the switch is here so its obvious where it will live */}
          <Tooltip delayDuration={100}>
            <TooltipTrigger asChild>
              <span>
                <Switch
                  aria-label='Automatically spend charges on this schedule'
                  checked={false}
                  data-testid='recurring-auto-spend'
                  disabled
                />
              </span>
            </TooltipTrigger>
            <TooltipContent side='top'>Automatically spending new charges is coming soon.</TooltipContent>
          </Tooltip>
        </div>
        <div className={styles.controls}>
          <Select
            className={styles.select}
            disabled={saving || !hasExpenses}
            isLoading={spendingIsLoading || balancesIsLoading}
            onChange={linkSpending}
            optionComponent={SelectSpendingOptionComponent}
            options={options}
            placeholder={placeholder}
            value={hasExpenses ? value : undefined}
          />
          {!props.recurring.spendingId && !props.recurring.ended && (
            <Button
              className={styles.createButton}
              data-testid='recurring-new-expense'
              disabled={saving}
              onClick={() =>
                showNewExpenseModal({
                  recurring: props.recurring,
                })
              }
              variant='secondary'
            >
              <Plus />
              New Expense
            </Button>
          )}
        </div>
        {props.recurring.spending && props.recurring.autoMatched && (
          <span className={styles.note} data-testid='recurring-auto-matched'>
            <Sparkles />
            monetr picked {props.recurring.spending.name} for you since your recent charges were spent from it.
          </span>
        )}
        {props.recurring.spending && (
          <Link
            className={styles.viewLink}
            data-testid='recurring-view-expense'
            to={`/bank/${props.recurring.bankAccountId}/expenses/${props.recurring.spending.spendingId}/details`}
          >
            View Expense
          </Link>
        )}
      </div>
    </section>
  );
}
