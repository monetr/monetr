import { Fragment, useCallback, useId, useRef } from 'react';
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
import MSpan from '@monetr/interface/components/MSpan';
import { Switch } from '@monetr/interface/components/Switch';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/useCreateFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewFundingValues {
  name: string;
  nextOccurrence: Date;
  ruleset: string;
  excludeWeekends: boolean;
  estimatedDeposit?: number | null;
}

function NewFundingModal(): JSX.Element {
  const switchId = useId();
  const { inTimezone } = useTimezone();
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();
  const {
    data: { friendlyToAmount },
  } = useLocaleCurrency();

  const initialValues: NewFundingValues = {
    name: '',
    nextOccurrence: startOfTomorrow({
      in: inTimezone,
    }),
    ruleset: '',
    excludeWeekends: false,
    estimatedDeposit: undefined,
  };

  const submit = useCallback(
    async (values: NewFundingValues, helpers: FormikHelpers<NewFundingValues>): Promise<void> => {
      helpers.setSubmitting(true);
      return await createFundingSchedule({
        bankAccountId: selectedBankAccountId,
        name: values.name,
        nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
          in: inTimezone,
        }),
        ruleset: values.ruleset,
        estimatedDeposit: values.estimatedDeposit > 0 ? friendlyToAmount(values.estimatedDeposit) : null,
        excludeWeekends: values.excludeWeekends,
      })
        .then(created => modal.resolve(created))
        .then(() => modal.remove())
        .catch(
          (error: AxiosError<APIError>) =>
            void enqueueSnackbar(error.response.data.error, {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [createFundingSchedule, enqueueSnackbar, friendlyToAmount, modal, selectedBankAccountId, inTimezone],
  );

  return (
    <MModal className='md:max-w-md' open={modal.visible} ref={ref}>
      <MForm
        className='h-full flex flex-col gap-2 p-2 justify-between'
        data-testid='new-funding-modal'
        initialValues={initialValues}
        onSubmit={submit}
      >
        {({ setFieldValue, values }) => (
          <Fragment>
            <div className='flex flex-col'>
              <MSpan className='font-bold text-xl mb-2'>Create A New Funding Schedule</MSpan>
              <FormTextField
                autoComplete='off'
                autoFocus
                data-1p-ignore
                label='What do you want to call your funding schedule?'
                name='name'
                placeholder='Example: Payday...'
                required
              />
              <FormDatePicker
                label='When do you get paid next?'
                min={startOfTomorrow({
                  in: inTimezone,
                })}
                name='nextOccurrence'
                required
              />
              <MSelectFrequency
                dateFrom='nextOccurrence'
                label='How often do you get paid?'
                name='ruleset'
                placeholder='Select a funding frequency...'
                required
              />
              <FormAmountField
                allowNegative={false}
                label='Estimated Deposit'
                name='estimatedDeposit'
                placeholder='Example: $ 1,000.00'
              />
              <div className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string mb-4'>
                <div className='space-y-0.5'>
                  <label
                    className='text-sm font-medium text-dark-monetr-content-emphasis cursor-pointer'
                    htmlFor={switchId}
                  >
                    Exclude Weekends
                  </label>
                  <p className='text-sm text-dark-monetr-content'>
                    If it were to land on a weekend, it is adjusted to the previous weekday instead.
                  </p>
                </div>
                <Switch
                  checked={values.excludeWeekends}
                  id={switchId}
                  onCheckedChange={() => setFieldValue('excludeWeekends', !values.excludeWeekends)}
                />
              </div>
            </div>
            <div className='flex justify-end gap-2'>
              <FormButton data-testid='close-new-funding-modal' onClick={modal.remove} variant='destructive'>
                Cancel
              </FormButton>
              <FormButton type='submit' variant='primary'>
                Create
              </FormButton>
            </div>
          </Fragment>
        )}
      </MForm>
    </MModal>
  );
}

const newFundingModal = NiceModal.create(NewFundingModal);

export default newFundingModal;

export function showNewFundingModal(): Promise<FundingSchedule | null> {
  return NiceModal.show<FundingSchedule | null, ExtractProps<typeof newFundingModal>, unknown>(newFundingModal);
}
