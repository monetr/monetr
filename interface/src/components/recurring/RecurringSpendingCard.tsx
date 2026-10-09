import { useState } from 'react';
import { ArrowRight, Plus, Repeat, Sparkles } from 'lucide-react';
import { rrulestr } from 'rrule';
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
import { useSpending } from '@monetr/interface/hooks/useSpending';
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
  // The recurring transaction only has the id, look the expense up so it comes from the same cache as the expenses
  const { data: linkedSpending } = useSpending(props.recurring.spendingId);
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

  async function toggleAutoAssign(autoAssign: boolean) {
    setSaving(true);
    return await patchTransactionRecurring({
      transactionRecurringId: props.recurring.transactionRecurringId,
      bankAccountId: props.recurring.bankAccountId,
      autoAssign,
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

  let autoAssignTooltip = 'Automatically spend new charges from this expense when they show up.';
  if (!props.recurring.spendingId) {
    autoAssignTooltip = 'Pick an expense to automatically spend new charges from it.';
  }

  return (
    <section className={styles.root}>
      <div className={styles.header}>
        <Typography color='emphasis' component='h3' size='md' weight='semibold'>
          Where New Charges Go
        </Typography>
        <Typography color='subtle' size='sm'>
          Pick where new {props.name || 'recurring'} charges come out of when they show up.
        </Typography>
      </div>
      <div className={styles.row}>
        <div className={styles.rowName}>
          <span className={styles.rowIcon}>
            <Repeat />
          </span>
          <div className={styles.rowText}>
            <span className={styles.rowTitle}>{props.name || 'Recurring'} Charges</span>
            <span className={styles.rowDetail}>
              {locale?.formatAmount(Math.abs(props.recurring.lastAmount), AmountType.Stored)}{' '}
              {rrulestr(props.recurring.ruleset).toText()}
            </span>
          </div>
        </div>
        <ArrowRight className={styles.arrow} />
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
          <Tooltip delayDuration={100}>
            <TooltipTrigger asChild>
              <span>
                <Switch
                  aria-label='Automatically spend charges on this schedule'
                  checked={props.recurring.autoAssign}
                  data-testid='recurring-auto-assign'
                  disabled={saving || !props.recurring.spendingId}
                  onCheckedChange={toggleAutoAssign}
                />
              </span>
            </TooltipTrigger>
            <TooltipContent side='top'>{autoAssignTooltip}</TooltipContent>
          </Tooltip>
        </div>
      </div>
      {props.recurring.spendingId && (
        <div className={styles.footer}>
          {props.recurring.autoMatched && linkedSpending && (
            <span className={styles.note} data-testid='recurring-auto-matched'>
              <Sparkles />
              monetr picked {linkedSpending.name} for you since your recent charges were spent from it.
            </span>
          )}
          <Link
            className={styles.viewLink}
            data-testid='recurring-view-expense'
            to={`/bank/${props.recurring.bankAccountId}/expenses/${props.recurring.spendingId}/details`}
          >
            View Expense
          </Link>
        </div>
      )}
      {!props.recurring.spendingId && !props.recurring.ended && (
        <div className={styles.footer}>
          <Button
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
        </div>
      )}
    </section>
  );
}
