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

import styles from './RecurringFundingCard.module.scss';

export interface RecurringFundingCardProps {
  recurring: TransactionRecurring;
  name: string;
}

// Same idea as the spending card but for money coming in, like a paycheck funding a schedule
export default function RecurringFundingCard(props: RecurringFundingCardProps): React.JSX.Element {
  const { data: fundingSchedules, isLoading } = useFundingSchedules();
  const patchTransactionRecurring = usePatchTransactionRecurring();
  const { enqueueSnackbar } = useSnackbar();
  const [saving, setSaving] = useState(false);

  const options: Array<SelectOption<ID<FundingSchedule> | null>> = [
    {
      label: 'Nothing',
      value: null,
    },
    ...(fundingSchedules ?? [])
      .sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
      .map(item => ({
        label: item.name,
        value: item.fundingScheduleId,
      })),
  ];
  const hasFundingSchedules = (fundingSchedules?.length ?? 0) > 0;
  const value = options.find(item => item.value === props.recurring.fundingScheduleId);

  let placeholder = 'Select a funding schedule...';
  if (!hasFundingSchedules) {
    placeholder = 'No funding schedules exist...';
  }

  async function linkFundingSchedule(newValue: SelectOption<ID<FundingSchedule> | null>) {
    if (newValue.value === props.recurring.fundingScheduleId) {
      return;
    }

    setSaving(true);
    return await patchTransactionRecurring({
      transactionRecurringId: props.recurring.transactionRecurringId,
      bankAccountId: props.recurring.bankAccountId,
      fundingScheduleId: newValue.value,
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
          Funds
        </Typography>
        <Typography color='subtle' size='sm'>
          The funding schedule {props.name || 'these'} deposits go towards.
        </Typography>
      </div>
      <div className={styles.card} data-active={String(Boolean(props.recurring.fundingSchedule))}>
        <div className={styles.controls}>
          <Select
            className={styles.select}
            disabled={saving || !hasFundingSchedules}
            isLoading={isLoading}
            onChange={linkFundingSchedule}
            options={options}
            placeholder={placeholder}
            value={hasFundingSchedules ? value : undefined}
          />
          {!props.recurring.fundingSchedule && !props.recurring.ended && (
            <Button
              className={styles.createButton}
              disabled={saving}
              onClick={() =>
                showNewFundingModal({
                  recurring: props.recurring,
                })
              }
              variant='secondary'
            >
              <Plus />
              New Funding Schedule
            </Button>
          )}
        </div>
        {props.recurring.fundingSchedule && (
          <Link
            className={styles.viewLink}
            to={`/bank/${props.recurring.bankAccountId}/funding/${props.recurring.fundingSchedule.fundingScheduleId}/details`}
          >
            View Funding Schedule
          </Link>
        )}
      </div>
    </section>
  );
}
