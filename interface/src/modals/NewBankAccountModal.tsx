import React, { useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { AxiosError } from 'axios';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import FormButton from '@monetr/interface/components/FormButton';
import MAmountField from '@monetr/interface/components/MAmountField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import { useCreateBankAccount, useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useAuthenticationSink } from '@monetr/interface/hooks/useAuthentication';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { friendlyToAmount } from '@monetr/interface/util/amounts';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewBankAccountValues {
  name: string;
  balance: number;
}

const initialValues: NewBankAccountValues = {
  name: '',
  balance: 0,
};

function NewBankAccountModal(): JSX.Element {
  const { data } = useAuthenticationSink();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const createBankAccount = useCreateBankAccount();
  const navigate = useNavigate();
  const ref = useRef<MModalRef>(null);

  const submit = useCallback(async (
    values: NewBankAccountValues,
    helper: FormikHelpers<NewBankAccountValues>,
  ): Promise<void> => {
    helper.setSubmitting(true);
    return await createBankAccount({
      linkId: selectedBankAccount.linkId,
      name: values.name,
      availableBalance: friendlyToAmount(values.balance),
      currentBalance: friendlyToAmount(values.balance),
      // TODO Make it so these can be customized
      accountType: BankAccountType.Depository,
      accountSubType: BankAccountSubType.Checking,
    })
      .then(result => navigate(`/bank/${ result.bankAccountId }/transactions`))
      .then(() => modal.remove())
      .catch((error: AxiosError) => void enqueueSnackbar(error.response.data['error'], {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helper.setSubmitting(false));
  }, [createBankAccount, selectedBankAccount, navigate, modal, enqueueSnackbar]);

  return (
    <MModal open={ modal.visible } ref={ ref }>
      <MForm
        onSubmit={ submit }
        initialValues={ initialValues }
        className='h-full flex flex-col gap-2 p-2 justify-between'
        data-testid='new-bank-account-modal'
      >
        <div className='flex flex-col'>
          <MSpan weight='bold' size='xl' className='mb-2'>
            Create A New Bank Account
          </MSpan>
          <MTextField
            id='bank-account-name-search' // Keep's 1Pass from hijacking normal name fields.
            data-testid='bank-account-name'
            name='name'
            label={ 'What is the account\'s name ?' }
            required
            autoFocus
            autoComplete='off'
            placeholder='Personal Checking...'
          />
          <MAmountField
            id='bank-account-balance-search' // Keep's 1Pass from hijacking normal name fields.
            data-testid='bank-account-balance'
            name='balance'
            label='Initial Balance'
            required
            currency={ data.defaultCurrency }
          />
        </div>
        <div className='flex justify-end gap-2'>
          <FormButton variant='secondary' onClick={ modal.remove } data-testid='close-new-bank-account-modal'>
            Cancel
          </FormButton>
          <FormButton variant='primary' type='submit' data-testid='bank-account-submit'>
            Create
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const newBankAccountModal = NiceModal.create(NewBankAccountModal);

export default newBankAccountModal;

export function showNewBankAccountModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof newBankAccountModal>, Object>(newBankAccountModal);
}
