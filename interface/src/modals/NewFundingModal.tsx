import React, { Fragment, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { AxiosError } from 'axios';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import MAmountField from '@monetr/interface/components/MAmountField';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { Switch } from '@monetr/interface/components/Switch';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewFundingValues {
  name: string;
  nextOccurrence: Date;
  ruleset: string;
  excludeWeekends: boolean;
  estimatedDeposit?: number | null;
}

const initialValues: NewFundingValues = {
  name: '',
  nextOccurrence: startOfTomorrow(),
  ruleset: '',
  excludeWeekends: false,
  estimatedDeposit: undefined,
};

function NewFundingModal(): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();
  const { data: { friendlyToAmount } } = useLocaleCurrency();

  async function submit(values: NewFundingValues, helpers: FormikHelpers<NewFundingValues>): Promise<void> {
    helpers.setSubmitting(true);
    const newFundingSchedule = new FundingSchedule({
      bankAccountId: selectedBankAccountId,
      name: values.name,
      nextRecurrence: startOfDay(new Date(values.nextOccurrence)),
      ruleset: values.ruleset,
      estimatedDeposit: values.estimatedDeposit > 0 ? friendlyToAmount(values.estimatedDeposit) : null,
      excludeWeekends: values.excludeWeekends,
    });

    return createFundingSchedule(newFundingSchedule)
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch((error: AxiosError) => void enqueueSnackbar(error.response.data['error'], {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helpers.setSubmitting(false));
  }

  return (
    <MModal open={ modal.visible } ref={ ref } className='md:max-w-md'>
      <MForm
        initialValues={ initialValues }
        onSubmit={ submit }
        className='h-full flex flex-col gap-2 p-2 justify-between' data-testid='new-funding-modal'
      >
        { ({ setFieldValue, values }) => (
          <Fragment>
            <div className='flex flex-col'>
              <MSpan className='font-bold text-xl mb-2'>
                Create A New Funding Schedule
              </MSpan>
              <MTextField
                autoFocus
                id='funding-name-search' // Keep's 1Pass from hijacking normal name fields.
                name='name'
                label='What do you want to call your funding schedule?'
                required
                autoComplete='off'
                placeholder='Example: Payday...'
              />
              <MDatePicker
                name='nextOccurrence'
                label='When do you get paid next?'
                required
                min={ startOfTomorrow() }
              />
              <MSelectFrequency
                dateFrom='nextOccurrence'
                menuPosition='fixed'
                menuShouldScrollIntoView={ false }
                menuShouldBlockScroll={ true }
                menuPortalTarget={ document.body }
                menuPlacement='bottom'
                label='How often do you get paid?'
                placeholder='Select a funding frequency...'
                required
                name='ruleset'
              />
              <MAmountField
                allowNegative={ false }
                label='Estimated Deposit'
                name='estimatedDeposit'
                placeholder='Example: $ 1,000.00'
              />
              <div className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string mb-4'>
                <div className='space-y-0.5'>
                  <label className='text-sm font-medium text-dark-monetr-content-emphasis cursor-pointer'>
                    Exclude Weekends
                  </label>
                  <p className='text-sm text-dark-monetr-content'>
                    If it were to land on a weekend, it is adjusted to the previous weekday instead.
                  </p>
                </div>
                <Switch
                  checked={ values['excludeWeekends'] }
                  onCheckedChange={ () => setFieldValue('excludeWeekends', !values['excludeWeekends']) }
                />
              </div>
            </div>
            <div className='flex justify-end gap-2'>
              <FormButton variant='destructive' onClick={ modal.remove } data-testid='close-new-funding-modal'>
                Cancel
              </FormButton>
              <FormButton variant='primary' type='submit'>
                Create
              </FormButton>
            </div>
          </Fragment>
        ) }
      </MForm>
    </MModal>
  );
}

const newFundingModal = NiceModal.create(NewFundingModal);

export default newFundingModal;

export function showNewFundingModal(): Promise<FundingSchedule | null> {
  return NiceModal.show<FundingSchedule | null, ExtractProps<typeof newFundingModal>, {}>(newFundingModal);
}
