import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Formik } from 'formik';
import moment from 'moment';

import MForm from 'components/MForm';
import MModal from 'components/MModal';
import Recurrence from 'components/Recurrence/Recurrence';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useFundingSchedules } from 'hooks/fundingSchedules';
import { useCreateSpending } from 'hooks/spending';
import MFormButton from 'components/MButton';
import MTextField from 'components/MTextField';
import { TextInput } from '@tremor/react';
import MSpan from 'components/MSpan';

import Select, { ActionMeta, components, FormatOptionLabelMeta, OnChangeValue, OptionProps, Theme } from 'react-select';
import useTheme from 'hooks/useTheme';
import MSelect from 'components/MSelect';


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
  const ref = useRef<HTMLDivElement>(null);
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
            For your Flagship Checking at Navy Federal Credit Union
          </MSpan>
          <div className='flex flex-col'>
            <MTextField
              label='What are you budgeting for?'
              required
              placeholder='Amazon, Netflix...'
            />
            <div className='flex gap-4'>
              <MTextField
                label='How much do you need?'
                required
                type='number'
                className='w-1/2'
              />
              <MTextField
                label='When do you need it next?'
                required
                type='date'
                className='w-1/2'
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
            />
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
              label='How frequently do you need this expense?'
              placeholder='Select a spending frequency...'
              required
            />
          </div>


          <div className='flex justify-end gap-2'>
            <MFormButton color='cancel' onClick={ modal.remove }>
              Cancel
            </MFormButton>
            <MFormButton color='primary'>
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
