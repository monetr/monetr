import { useState } from 'react';
import { Plus } from 'lucide-react';
import { Link } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import Typography from '@monetr/interface/components/Typography';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { usePatchTransactionRecurring } from '@monetr/interface/hooks/usePatchTransactionRecurring';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './RecurringLinkCard.module.scss';

export interface RecurringFundingCardProps {
  recurring: TransactionRecurring;
  name: string;
}

// Same idea as the spending card but for money coming in, like a paycheck funding a schedule.
export default function RecurringFundingCard({ recurring, name }: RecurringFundingCardProps): React.JSX.Element {
  const { data: fundingSchedules, isLoading } = useFundingSchedules();
  const patchTransactionRecurring = usePatchTransactionRecurring();
  const { enqueueSnackbar } = useSnackbar();
  const [saving, setSaving] = useState(false);

  const options: Array<SelectOption<ID<FundingSchedule> | null>> = [
    { label: 'Nothing', value: null },
    ...(fundingSchedules ?? [])
      .sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
      .map(item => ({ label: item.name, value: item.fundingScheduleId })),
  ];
  const hasFundingSchedules = (fundingSchedules?.length ?? 0) > 0;
  const value = options.find(option => option.value === recurring.fundingScheduleId);

  async function onChange(newValue: SelectOption<ID<FundingSchedule> | null>) {
    if (newValue.value === recurring.fundingScheduleId) {
      return;
    }

    setSaving(true);
    return await patchTransactionRecurring({
      transactionRecurringId: recurring.transactionRecurringId,
      bankAccountId: recurring.bankAccountId,
      fundingScheduleId: newValue.value,
    })
      .then(
        () =>
          void enqueueSnackbar(newValue.value ? `Linked to ${newValue.label}` : 'No longer funding anything', {
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
          Funds
        </Typography>
        <Typography color='subtle' size='sm'>
          The funding schedule {name || 'these'} deposits go towards.
        </Typography>
      </div>
      <div className={styles.card} data-active={String(Boolean(recurring.fundingSchedule))}>
        <div className={styles.controls}>
          <Select
            className={styles.select}
            disabled={saving || !hasFundingSchedules}
            isLoading={isLoading}
            onChange={onChange}
            options={options}
            placeholder={hasFundingSchedules ? 'Select a funding schedule...' : 'No funding schedules exist...'}
            value={hasFundingSchedules ? value : undefined}
          />
          {!recurring.fundingSchedule && !recurring.ended && (
            <Button
              className={styles.createButton}
              disabled={saving}
              onClick={() => showNewFundingModal({ recurring })}
              variant='secondary'
            >
              <Plus />
              New funding schedule
            </Button>
          )}
        </div>
        {recurring.fundingSchedule && (
          <Link
            className={styles.viewLink}
            to={`/bank/${recurring.bankAccountId}/funding/${recurring.fundingSchedule.fundingScheduleId}/details`}
          >
            View funding schedule
          </Link>
        )}
      </div>
    </section>
  );
}
