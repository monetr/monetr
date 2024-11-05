import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay } from 'date-fns';

import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewTransactionValues {
  name: string;
  date: Date;
  spendingId: string | null;
  amount: number;
  kind: 'debit' | 'credit';
  // TODO Just keep this false for now, monetr does not allow pending to be modified.
  isPending: boolean;
  adjustsBalance: boolean;
}

const initialValues: NewTransactionValues = {
  name: '',
  date: startOfDay(new Date()),
  amount: 0,
  spendingId: null,
  kind: 'debit',
  isPending: false,
  adjustsBalance: false,
};

function NewTransactionModal(): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);

  return (
    <MModal open={ modal.visible } ref={ ref } className='sm:max-w-xl'>
      <MForm
        onSubmit={ () => {} }
        className='h-full flex flex-col gap-2 p-2 justify-between'
        initialValues={ initialValues }
      >
        <div className='flex flex-col'>
          <MSpan weight='bold' size='xl' className='mb-2'>
            Create A New Transaction
          </MSpan>
        </div>
        <div className='flex justify-end gap-2'>
          <FormButton variant='destructive' onClick={ modal.remove } data-testid='close-new-transaction-modal'>
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

const newTransactionModal = NiceModal.create(NewTransactionModal);

export default newTransactionModal;

export function showNewTransactionModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof newTransactionModal>, {}>(newTransactionModal);
}
