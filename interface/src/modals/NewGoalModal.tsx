import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import Typography from '@monetr/interface/components/Typography';
import { useCreateSpending } from '@monetr/interface/hooks/useCreateSpending';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import { SpendingType } from '@monetr/interface/models/Spending';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './NewGoalModal.module.scss';

interface NewGoalValues {
  name: string;
  amount: number;
  nextOccurrence: Date;
  fundingScheduleId: ID<FundingSchedule>;
}

function NewGoalModal(): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createSpending = useCreateSpending();
  const ref = useRef<MModalRef>(null);

  const initialValues: NewGoalValues = {
    name: '',
    amount: 0.0,
    nextOccurrence: startOfTomorrow({
      in: inTimezone,
    }),
    fundingScheduleId: ID.from<FundingSchedule, string>(''),
  };

  async function submit(values: NewGoalValues, helper: FormikHelpers<NewGoalValues>): Promise<void> {
    if (!locale || !selectedBankAccountId) {
      return Promise.resolve();
    }

    helper.setSubmitting(true);
    return await createSpending({
      bankAccountId: selectedBankAccountId,
      name: values.name.trim(),
      nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
        in: inTimezone,
      }),
      spendingType: SpendingType.Goal,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: locale.friendlyToAmount(values.amount),
      ruleset: null,
    })
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} data-testid='new-goal-modal' initialValues={initialValues} onSubmit={submit}>
        <div className={styles.body}>
          <Typography className={styles.heading} size='xl' weight='bold'>
            Create A New Goal
          </Typography>
          <FormTextField
            autoComplete='off'
            autoFocus
            data-1p-ignore
            label='What are you budgeting for?'
            name='name'
            placeholder='Vacation, Furniture, Car...'
            required
          />
          <div className={styles.fieldRow}>
            <FormAmountField
              allowNegative={false}
              className={styles.fieldRowItem}
              label='How much do you need?'
              name='amount'
              required
            />
            <FormDatePicker
              className={styles.fieldRowItem}
              label='How soon will you need it?'
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='nextOccurrence'
              required
            />
          </div>
          <MSelectFunding
            label='When do you want to fund the goal?'
            menuPortalTarget={document.body}
            name='fundingScheduleId'
            required
          />
        </div>
        <div className={styles.actions}>
          <Button data-testid='close-new-goal-modal' onClick={modal.remove} variant='destructive'>
            Cancel
          </Button>
          <FormButton type='submit' variant='primary'>
            Create
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const newGoalModal = NiceModal.create(NewGoalModal);

export default newGoalModal;

export function showNewGoalModal(): Promise<Spending | null> {
  return NiceModal.show(newGoalModal) as Promise<Spending | null>;
}
