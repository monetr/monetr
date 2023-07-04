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
import MButton from 'components/MButton';
import MTextField from 'components/MTextField';
import { TextInput } from '@tremor/react';

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
  const selectedBankAccountId = useSelectedBankAccountId();
  const createSpending = useCreateSpending();
  const fundingSchedules = useFundingSchedules();

  const ref = useRef<HTMLDivElement>(null);

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
          <div className='flex flex-col gap-2'>
            <TextInput placeholder='Name of the expense' />
          </div>
          <div>
            <MButton color='primary'>
              Create
            </MButton>
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
