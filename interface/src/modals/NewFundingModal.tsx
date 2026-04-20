import { Fragment, useCallback, useId, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { type FormikHelpers, useFormikContext } from 'formik';
import { useSnackbar } from 'notistack';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import { Switch } from '@monetr/interface/components/Switch';
import Typography from '@monetr/interface/components/Typography';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/useCreateFundingSchedule';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
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
  autoCreateTransaction: boolean;
}

function NewFundingModal(): JSX.Element {
  const switchId = useId();
  const { inTimezone } = useTimezone();
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();
  const { data: link } = useCurrentLink();
  const isManual = Boolean(link?.getIsManual());
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
    autoCreateTransaction: false,
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
        // Auto create transaction requires a manual link and a non-zero
        // estimated deposit; force it off otherwise so the API will not reject
        // the create.
        autoCreateTransaction: isManual && (values.estimatedDeposit ?? 0) > 0 && values.autoCreateTransaction,
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
        .finally(() => helpers.setSubmitting(false));
    },
    [createFundingSchedule, enqueueSnackbar, friendlyToAmount, modal, selectedBankAccountId, inTimezone, isManual],
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
              <Typography className='mb-2' size='xl' weight='bold'>
                Create A New Funding Schedule
              </Typography>
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
              {isManual && <AutoCreateTransactionToggle />}
            </div>
            <div className='flex justify-end gap-2'>
              <Button data-testid='close-new-funding-modal' onClick={modal.remove} variant='secondary'>
                Cancel
              </Button>
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

function AutoCreateTransactionToggle(): JSX.Element {
  const autoCreateSwitchId = useId();
  const { setFieldValue, values } = useFormikContext<NewFundingValues>();
  const hasDeposit = (values.estimatedDeposit ?? 0) > 0;

  return (
    <div
      className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string mb-4'
      data-testid='new-funding-auto-create-transaction'
    >
      <div className='space-y-0.5'>
        <label
          aria-disabled={!hasDeposit}
          className='text-sm font-medium text-dark-monetr-content-emphasis cursor-pointer aria-disabled:cursor-not-allowed aria-disabled:opacity-50'
          htmlFor={autoCreateSwitchId}
        >
          Auto create transaction
        </label>
        <p aria-disabled={!hasDeposit} className='text-sm text-dark-monetr-content aria-disabled:opacity-50'>
          Automatically add a deposit transaction for the estimated deposit each time the funding schedule would occur.
        </p>
      </div>
      <Switch
        checked={hasDeposit && values.autoCreateTransaction}
        disabled={!hasDeposit}
        id={autoCreateSwitchId}
        onCheckedChange={() => setFieldValue('autoCreateTransaction', !values.autoCreateTransaction)}
      />
    </div>
  );
}

const newFundingModal = NiceModal.create(NewFundingModal);

export default newFundingModal;

export function showNewFundingModal(): Promise<FundingSchedule | null> {
  return NiceModal.show<FundingSchedule | null, ExtractProps<typeof newFundingModal>, unknown>(newFundingModal);
}
