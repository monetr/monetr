import React, { Fragment, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfToday } from 'date-fns';

import FormButton from '@monetr/interface/components/FormButton';
import MAmountField from '@monetr/interface/components/MAmountField';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@monetr/interface/components/Tabs';
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
  date: startOfToday(),
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
        { ({ setFieldValue }) => (
          <Fragment>
            <div className='flex flex-col'>
              <MSpan weight='bold' size='xl' className='mb-2'>
                  Create A New Transaction
              </MSpan>

              { /* 
              TODO I'm like 99% sure there is going to be a bug here where someone could do something like select a
              spending ID while on the debit tab, then switch to the credit tab and create a deposit with a spending ID?
              */ }

              <Tabs 
                defaultValue='debit' 
                className='w-full mb-2'
                onValueChange={ value => setFieldValue('kind', value as any) }
              >
                <TabsList className='w-full'>
                  <TabsTrigger className='w-full' value='debit'>Debit</TabsTrigger>
                  <TabsTrigger className='w-full' value='credit'>Credit</TabsTrigger>
                </TabsList>
                <TabsContent value='debit'>
                  <MTextField
                    name='name'
                    label='Name / Description'
                    required
                    autoComplete='off'
                    placeholder='Amazon, Netflix...'
                    data-1p-ignore 
                  />
                  <div className='flex gap-0 md:gap-4 flex-col md:flex-row'>
                    <MAmountField
                      name='amount'
                      label='Amount'
                      required
                      className='w-full md:w-1/2'
                      allowNegative={ false }
                    />
                    <MDatePicker
                      className='w-full md:w-1/2'
                      label='Date'
                      min={ startOfToday() }
                      name='date'
                      required
                    />
                  </div>
                  <MSelectSpending
                    className='w-full'
                    name='spendingId'
                    menuPosition='fixed'
                    menuShouldScrollIntoView={ false }
                    menuShouldBlockScroll={ true }
                    menuPortalTarget={ document.body }
                    menuPlacement='bottom'
                  />
                </TabsContent>
                <TabsContent value='credit'> 
                  <MTextField
                    name='name'
                    label='Name / Description'
                    required
                    autoComplete='off'
                    placeholder='Paycheck, Deposit...'
                    data-1p-ignore 
                  />
                  <div className='flex gap-0 md:gap-4 flex-col md:flex-row'>
                    <MAmountField
                      name='amount'
                      label='Amount'
                      required
                      className='w-full md:w-1/2'
                      allowNegative={ false }
                    />
                    <MDatePicker
                      className='w-full md:w-1/2'
                      label='Date'
                      min={ startOfToday() }
                      name='date'
                      required
                    />
                  </div>
                </TabsContent>
              </Tabs>

            </div>
            <div className='flex justify-end gap-2'>
              <FormButton variant='destructive' onClick={ modal.remove } data-testid='close-new-transaction-modal'>
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

const newTransactionModal = NiceModal.create(NewTransactionModal);

export default newTransactionModal;

export function showNewTransactionModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof newTransactionModal>, {}>(newTransactionModal);
}
