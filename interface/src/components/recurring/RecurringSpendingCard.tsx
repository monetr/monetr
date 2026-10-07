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

import styles from './RecurringLinkCard.module.scss';

type SpendingOption = Pick<Spending | FreeToUse, 'spendingId' | 'spendingType' | 'currentAmount' | 'name'>;

export interface RecurringSpendingCardProps {
  recurring: TransactionRecurring;
  name: string;
}

export default function RecurringSpendingCard({ recurring, name }: RecurringSpendingCardProps): React.JSX.Element {
  const { data: spending, isLoading: spendingIsLoading } = useSpendings();
  const { data: balances, isLoading: balancesIsLoading } = useCurrentBalance();
  const { data: locale } = useLocaleCurrency();
  const patchTransactionRecurring = usePatchTransactionRecurring();
  const { enqueueSnackbar } = useSnackbar();
  const [saving, setSaving] = useState(false);

  // Free-To-Use is how you say its not budgeted with anything, same as everywhere else spending gets picked. Only
  // expenses can be linked for now, goals will get their own thing later.
  const options: Array<SelectOption<SpendingOption>> = [];
  if (balances) {
    options.push({ label: 'Free-To-Use', value: new FreeToUse(balances) });
  }
  const expenses = (spending ?? [])
    .filter(item => item.spendingType === SpendingType.Expense)
    .sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1));
  for (const expense of expenses) {
    options.push({ label: expense.name, value: expense });
  }

  const hasExpenses = options.some(option => option.value.spendingId !== FREE_TO_USE);
  const value = options.find(option => option.value.spendingId === (recurring.spendingId ?? FREE_TO_USE));

  async function onChange(newValue: SelectOption<SpendingOption>) {
    // Picking Free-To-Use means clearing the link, so send null to the server.
    let spendingId: ID<Spending> | null = newValue.value.spendingId;
    if (spendingId === FREE_TO_USE) {
      spendingId = null;
    }

    if (spendingId === recurring.spendingId) {
      return;
    }

    setSaving(true);
    return await patchTransactionRecurring({
      transactionRecurringId: recurring.transactionRecurringId,
      bankAccountId: recurring.bankAccountId,
      spendingId,
    })
      .then(
        () =>
          void enqueueSnackbar(spendingId ? `Linked to ${newValue.label}` : 'No longer budgeted with an expense', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response?.data?.error || 'Failed to update the recurring transaction', {
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
          Automatically spend
        </Typography>
        <Typography color='subtle' size='sm'>
          Pick where new {name || 'recurring'} charges come out of when they show up.
        </Typography>
      </div>
      <div className={styles.card} data-active={String(Boolean(recurring.spendingId))}>
        <div className={styles.cardHeader}>
          <div className={styles.cardText}>
            <span className={styles.chip}>
              <Repeat />
              This schedule
            </span>
            <span className={styles.title}>Spend charges on this schedule from</span>
            <span className={styles.description}>
              Only the {locale?.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored)} charges monetr matches
              to this schedule.
            </span>
          </div>
          {/* Automatically spending isn't a thing yet, the switch is here so its obvious where it will live. */}
          <Tooltip delayDuration={100}>
            <TooltipTrigger asChild>
              <span>
                <Switch aria-label='Automatically spend charges on this schedule' checked={false} disabled />
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
            onChange={onChange}
            optionComponent={SelectSpendingOptionComponent}
            options={options}
            placeholder={hasExpenses ? 'Select an expense...' : 'No expenses exist...'}
            value={hasExpenses ? value : undefined}
          />
          {!recurring.spendingId && !recurring.ended && (
            <Button
              className={styles.createButton}
              disabled={saving}
              onClick={() => showNewExpenseModal({ recurring })}
              variant='secondary'
            >
              <Plus />
              New expense
            </Button>
          )}
        </div>
        {recurring.spending && recurring.autoMatched && (
          <span className={styles.note}>
            <Sparkles />
            monetr picked {recurring.spending.name} for you since your recent charges were spent from it.
          </span>
        )}
        {recurring.spending && (
          <Link
            className={styles.viewLink}
            to={`/bank/${recurring.bankAccountId}/expenses/${recurring.spending.spendingId}/details`}
          >
            View expense
          </Link>
        )}
      </div>
    </section>
  );
}
