import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { AxiosError } from 'axios';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MFormButton from '@monetr/interface/components/MButton';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
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

  async function submit(values: NewFundingValues, helpers: FormikHelpers<NewFundingValues>): Promise<void> {
    helpers.setSubmitting(true);
    const newFundingSchedule = new FundingSchedule({
      bankAccountId: selectedBankAccountId,
      name: values.name,
      nextOccurrence: startOfDay(new Date(values.nextOccurrence)),
      ruleset: values.ruleset,
      estimatedDeposit: null,
      excludeWeekends: false,
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
    <MModal open={ modal.visible } ref={ ref } className='py-4 md:max-w-sm'>
      <MForm
        initialValues={ initialValues }
        onSubmit={ submit }
        className='h-full flex flex-col gap-2 p-2 justify-between' data-testid='new-funding-modal'
      >
        <div className='flex flex-col'>
          <MSpan className='font-bold text-xl mb-2'>
            Create A New Funding Schedule
          </MSpan>
          <MTextField
            id='funding-name-search' // Keep's 1Pass from hijacking normal name fields.
            name='name'
            label='What do you want to call your funding schedule?'
            required
            autoComplete="off"
            placeholder="Example: Payday..."
          />
          <MDatePicker
            name='nextOccurrence'
            label='When do you get paid next?'
            required
            min={ startOfTomorrow() }
          />
          <MSelectFrequency
            dateFrom="nextOccurrence"
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
          <div className='flex justify-end gap-2'>
            <MFormButton color='cancel' onClick={ modal.remove } data-testid='close-new-funding-modal'>
              Cancel
            </MFormButton>
            <MFormButton color='primary' type='submit'>
              Create
            </MFormButton>
          </div>
        </div>
      </MForm>
    </MModal>
  );
}

const newFundingModal = NiceModal.create(NewFundingModal);

export default newFundingModal;

export function showNewFundingModal(): Promise<FundingSchedule | null> {
  return NiceModal.show<FundingSchedule | null, ExtractProps<typeof newFundingModal>, {}>(newFundingModal);
}
