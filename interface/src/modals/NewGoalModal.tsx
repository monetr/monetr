import { useRef } from 'react';
import { tz } from '@date-fns/tz';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { AxiosError } from 'axios';
import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MAmountField from '@monetr/interface/components/MAmountField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import MSpan from '@monetr/interface/components/MSpan';
import { useCreateSpending } from '@monetr/interface/hooks/useCreateSpending';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewGoalValues {
  name: string;
  amount: number;
  nextOccurrence: Date;
  fundingScheduleId: string;
}

function NewGoalModal(): JSX.Element {
  const { data: timezone } = useTimezone();
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
      in: tz(timezone),
    }),
    fundingScheduleId: '',
  };

  async function submit(values: NewGoalValues, helper: FormikHelpers<NewGoalValues>): Promise<void> {
    const newSpending = new Spending({
      bankAccountId: selectedBankAccountId,
      name: values.name.trim(),
      nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
        in: tz(timezone),
      }),
      spendingType: SpendingType.Goal,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: locale.friendlyToAmount(values.amount),
      ruleset: null,
    });

    helper.setSubmitting(true);
    return createSpending(newSpending)
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch(
        (error: AxiosError<APIError>) =>
          void enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal open={modal.visible} ref={ref} className='sm:max-w-xl'>
      <MForm
        onSubmit={submit}
        initialValues={initialValues}
        className='h-full flex flex-col gap-2 p-2 justify-between'
        data-testid='new-goal-modal'
      >
        <div className='flex flex-col'>
          <MSpan className='font-bold text-xl mb-2'>Create A New Goal</MSpan>
          <FormTextField
            name='name'
            label='What are you budgeting for?'
            required
            autoFocus
            autoComplete='off'
            placeholder='Vacation, Furniture, Car...'
            data-1p-ignore
          />
          <div className='flex gap-0 md:gap-4 flex-col md:flex-row'>
            <MAmountField
              name='amount'
              label='How much do you need?'
              required
              className='w-full md:w-1/2'
              allowNegative={false}
            />
            <FormDatePicker
              className='w-full md:w-1/2'
              label='How soon will you need it?'
              min={startOfTomorrow({
                in: tz(timezone),
              })}
              name='nextOccurrence'
              required
            />
          </div>
          <MSelectFunding
            menuPortalTarget={document.body}
            label='When do you want to fund the goal?'
            required
            name='fundingScheduleId'
          />
        </div>
        <div className='flex justify-end gap-2'>
          <FormButton variant='destructive' onClick={modal.remove} data-testid='close-new-goal-modal'>
            Cancel
          </FormButton>
          <FormButton variant='primary' type='submit'>
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
  return NiceModal.show<Spending | null, ExtractProps<typeof newGoalModal>, unknown>(newGoalModal);
}
