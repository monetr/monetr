import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Formik } from 'formik';
import moment from 'moment';

import MFormButton from 'components/MButton';
import MForm from 'components/MForm';
import MModal, { MModalRef } from 'components/MModal';
import MSelect from 'components/MSelect';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import Recurrence from 'components/Recurrence/Recurrence';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useFundingSchedules } from 'hooks/fundingSchedules';
import { useCreateSpending } from 'hooks/spending';
import useTheme from 'hooks/useTheme';
import MSelectFrequency from 'components/MSelectFrequency';


interface NewExpenseValues {
  name: string;
  amount: number;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
  fundingScheduleId: number;
}

const initialValues: NewExpenseValues = {

  name: '',
  amount: 0.00,
  nextOccurrence: moment().add(1, 'day'),
  recurrenceRule: new Recurrence(),
  fundingScheduleId: 0,
};

function NewExpenseModal(): JSX.Element {
  const modal = useModal();
  // const selectedBankAccountId = useSelectedBankAccountId();
  // const createSpending = useCreateSpending();
  const fundingSchedules = useFundingSchedules();
  //
  const ref = useRef<MModalRef>(null);
  const theme = useTheme();

  async function submit(): Promise<void> {
    return Promise.resolve();
  }

  return (
    <MModal open={ modal.visible } ref={ ref }>
      <Formik
        onSubmit={ submit }
        initialValues={ initialValues }
      >
        <MForm className='flex flex-col gap-2 p-2'>
          <h2 className='font-bold'>
            Create A New Expense
          </h2>
          <MSpan>
            For your Flagship Checking at place
          </MSpan>
          <div className='flex flex-col'>
            <MTextField
              name='name'
              label='What are you budgeting for?'
              required
              placeholder='Amazon, Netflix...'
            />
            <div className='flex gap-0 md:gap-4 flex-col md:flex-row'>
              <MTextField
                name='amount'
                label='How much do you need?'
                required
                type='number'
                className='w-full md:w-1/2'
              />
              <MTextField
                name='nextOccurrence'
                label='When do you need it next?'
                required
                type='date'
                className='w-full md:w-1/2'
              />
            </div>
            <MSelect
              options={ [
                {
                  label: 'Test',
                  value: 0,
                },
                {
                  label: 'Test Other',
                  value: 1,
                },
              ] }
              menuPortalTarget={ document.body }
              menuPlacement='auto'
              label='When do you want to fund the expense?'
              placeholder='Select a funding schedule...'
              required
              name='fundingScheduleId'
            />
            <MSelectFrequency
              dateFrom="nextOccurrence"
              // menuPosition='fixed'
              // menuShouldScrollIntoView={ false }
              // menuShouldBlockScroll={ true }
              menuPortalTarget={ ref.current }
              // menuIsOpen
              menuPlacement='auto'
              label='How frequently do you need this expense?'
              placeholder='Select a spending frequency...'
              required
              name='recurrenceRule'
            />
          </div>
          <div className='flex justify-end gap-2'>
            <MFormButton color='cancel' onClick={ modal.remove }>
              Cancel
            </MFormButton>
            <MFormButton color='primary' type='submit'>
              Create
            </MFormButton>
          </div>
        </MForm>
      </Formik>
    </MModal>
  );
}

const newExpenseModal = NiceModal.create(NewExpenseModal);

export default newExpenseModal;

export function showNewExpenseModal(): void {
  NiceModal.show(newExpenseModal);
}
