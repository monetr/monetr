import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { AxiosError } from 'axios';
import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import Typography from '@monetr/interface/components/Typography';
import { useCreateSpending } from '@monetr/interface/hooks/useCreateSpending';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewExpenseValues {
  name: string;
  amount: number;
  nextOccurrence: Date;
  ruleset: string;
  fundingScheduleId: string;
}

function NewExpenseModal(): JSX.Element {
  const { inTimezone } = useTimezone();
  const {
    data: { friendlyToAmount },
  } = useLocaleCurrency();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const createSpending = useCreateSpending();

  const ref = useRef<MModalRef>(null);

  const initialValues: NewExpenseValues = {
    name: '',
    amount: 0.0,
    nextOccurrence: startOfTomorrow({
      in: inTimezone,
    }),
    ruleset: '',
    fundingScheduleId: '',
  };

  async function submit(values: NewExpenseValues, helper: FormikHelpers<NewExpenseValues>): Promise<void> {
    const newSpending = new Spending({
      bankAccountId: selectedBankAccount.bankAccountId,
      name: values.name.trim(),
      nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
        in: inTimezone,
      }),
      spendingType: SpendingType.Expense,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: friendlyToAmount(values.amount),
      ruleset: values.ruleset,
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
    <MModal className='sm:max-w-xl' open={modal.visible} ref={ref}>
      <MForm
        className='h-full flex flex-col gap-2 p-2 justify-between'
        data-testid='new-expense-modal'
        initialValues={initialValues}
        onSubmit={submit}
      >
        <div className='flex flex-col'>
          <Typography className='mb-2' size='xl' weight='bold'>
            Create A New Expense
          </Typography>
          <FormTextField
            autoComplete='off'
            autoFocus
            data-1p-ignore
            label='What are you budgeting for?'
            name='name'
            placeholder='Amazon, Netflix...'
            required
          />
          <div className='flex gap-0 md:gap-4 flex-col md:flex-row'>
            <FormAmountField
              allowNegative={false}
              className='w-full md:w-1/2'
              label='How much do you need?'
              name='amount'
              required
            />
            <FormDatePicker
              className='w-full md:w-1/2'
              label='When do you need it next?'
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='nextOccurrence'
              required
            />
          </div>
          <MSelectFunding
            label='When do you want to fund the expense?'
            menuPortalTarget={document.body}
            name='fundingScheduleId'
            required
          />
          <MSelectFrequency
            dateFrom='nextOccurrence'
            label='How frequently do you need this expense?'
            name='ruleset'
            placeholder='Select a spending frequency...'
            required
          />
        </div>
        <div className='flex justify-end gap-2'>
          <FormButton data-testid='close-new-expense-modal' onClick={modal.remove} variant='destructive'>
            Cancel
          </FormButton>
          <FormButton type='submit' variant='primary'>
            Create
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const newExpenseModal = NiceModal.create(NewExpenseModal);

export default newExpenseModal;

export function showNewExpenseModal(): Promise<Spending | null> {
  return NiceModal.show<Spending | null, ExtractProps<typeof newExpenseModal>, unknown>(newExpenseModal);
}
