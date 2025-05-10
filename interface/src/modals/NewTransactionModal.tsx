import React, { Fragment, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { AxiosError } from 'axios';
import { tz } from '@date-fns/tz';
import { startOfDay, startOfToday } from 'date-fns';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import MAmountField from '@monetr/interface/components/MAmountField';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { Switch } from '@monetr/interface/components/Switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@monetr/interface/components/Tabs';
import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { CreateTransactionRequest, useCreateTransaction } from '@monetr/interface/hooks/transactions';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
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

function NewTransactionModal(): JSX.Element {
  const { data: timezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const createTransaction = useCreateTransaction();

  const initialValues: NewTransactionValues = {
    name: '',
    date: startOfToday({
      in: tz(timezone),
    }),
    amount: 0,
    spendingId: null,
    kind: 'debit',
    isPending: false,
    adjustsBalance: false,
  };

  async function submit(
    values: NewTransactionValues,
    helper: FormikHelpers<NewTransactionValues>,
  ): Promise<void> {
    const newTransactionRequest: CreateTransactionRequest = {
      bankAccountId: selectedBankAccount.bankAccountId,
      amount: locale.friendlyToAmount(
        values.kind === 'credit' ? values.amount * -1 : values.amount,
      ),
      name: values.name,
      merchantName: null,
      date: startOfDay(new Date(values.date), {
        in: tz(timezone),
      }),
      isPending: values.isPending,
      spendingId: values.spendingId,
      adjustsBalance: values.adjustsBalance,
    };

    helper.setSubmitting(true);

    return createTransaction(newTransactionRequest)
      // TODO Show toast that the transaction was created, include button to "view transaction".
      .then(() => modal.remove())
      .catch((error: AxiosError) => void enqueueSnackbar(error.response.data['error'], {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal open={ modal.visible } ref={ ref } className='sm:max-w-xl'>
      <MForm
        onSubmit={ submit }
        className='h-full flex flex-col gap-2 p-2 justify-between'
        initialValues={ initialValues }
      >
        { ({ setFieldValue, values }) => (
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
                    autoComplete='off'
                    autoFocus
                    data-1p-ignore 
                    label='Name / Description'
                    name='name'
                    placeholder='Amazon, Netflix...'
                    required
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
                  <div className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string mb-4'>
                    <div className='space-y-0.5'>
                      <label className='text-sm font-medium text-dark-monetr-content-emphasis cursor-pointer'>
                        Adjust Balance
                      </label>
                      <p className='text-sm text-dark-monetr-content'>
                        Update your account balance for this transaction?
                      </p>
                    </div>
                    <Switch
                      checked={ values['adjustsBalance'] }
                      onCheckedChange={ () => setFieldValue('adjustsBalance', !values['adjustsBalance']) }
                    />
                  </div>
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
                      name='date'
                      required
                    />
                  </div>
                  <div className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string mb-4'>
                    <div className='space-y-0.5'>
                      <label className='text-sm font-medium text-dark-monetr-content-emphasis'>
                        Adjust Balance
                      </label>
                      <p className='text-sm text-dark-monetr-content'>
                        Update your account balance for this transaction?
                      </p>
                    </div>
                    <Switch
                      checked={ values['adjustsBalance'] }
                      onCheckedChange={ () => setFieldValue('adjustsBalance', !values['adjustsBalance']) }
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
