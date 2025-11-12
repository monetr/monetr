import { Fragment, useCallback, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { AxiosError } from 'axios';
import type { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';
import { useNavigate } from 'react-router-dom';

import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import useLocaleCurrency, { DefaultCurrency } from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface NewBankAccountValues {
  name: string;
  balance: number;
  currency: string;
}

function NewBankAccountModal(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount, isLoading } = useSelectedBankAccount();
  const createBankAccount = useCreateBankAccount();
  const navigate = useNavigate();
  const ref = useRef<MModalRef>(null);

  const initialValues: NewBankAccountValues = {
    name: '',
    balance: 0,
    currency: locale?.currency ?? DefaultCurrency,
  };

  const submit = useCallback(
    async (values: NewBankAccountValues, helper: FormikHelpers<NewBankAccountValues>): Promise<void> => {
      helper.setSubmitting(true);
      return await createBankAccount({
        linkId: selectedBankAccount.linkId,
        name: values.name,
        availableBalance: locale.friendlyToAmount(values.balance),
        currentBalance: locale.friendlyToAmount(values.balance),
        // TODO Make it so these can be customized
        accountType: BankAccountType.Depository,
        accountSubType: BankAccountSubType.Checking,
        currency: values.currency,
      })
        .then(result => navigate(`/bank/${result.bankAccountId}/transactions`))
        .then(() => modal.remove())
        .catch(
          (error: AxiosError<APIError>) =>
            void enqueueSnackbar(error.response.data.error, {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        )
        .finally(() => helper.setSubmitting(false));
    },
    [createBankAccount, selectedBankAccount, locale, navigate, modal, enqueueSnackbar],
  );

  if (isLoading) {
    return (
      <MModal open={modal.visible} ref={ref}>
        One moment...
      </MModal>
    );
  }

  return (
    <MModal open={modal.visible} ref={ref}>
      <MForm
        className='h-full flex flex-col gap-2 p-2 justify-between'
        data-testid='new-bank-account-modal'
        initialValues={initialValues}
        onSubmit={submit}
      >
        {({ values }) => (
          <Fragment>
            <div className='flex flex-col'>
              <MSpan className='mb-2' size='xl' weight='bold'>
                Create A New Bank Account
              </MSpan>
              <FormTextField
                autoComplete='off'
                data-1p-ignore
                data-testid='bank-account-name'
                label="What is the account's name ?"
                name='name'
                placeholder='Personal Checking...'
                required
              />
              <SelectCurrency className='w-full' menuPortalTarget={document.body} name='currency' required />
              <FormAmountField
                allowNegative
                currency={values.currency}
                data-1p-ignore
                data-testid='bank-account-balance'
                label='Initial Balance'
                name='balance'
                required
              />
            </div>
            <div className='flex justify-end gap-2'>
              <FormButton data-testid='close-new-bank-account-modal' onClick={modal.remove} variant='secondary'>
                Cancel
              </FormButton>
              <FormButton data-testid='bank-account-submit' type='submit' variant='primary'>
                Create
              </FormButton>
            </div>
          </Fragment>
        )}
      </MForm>
    </MModal>
  );
}

const newBankAccountModal = NiceModal.create(NewBankAccountModal);

export default newBankAccountModal;

export function showNewBankAccountModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof newBankAccountModal>, unknown>(newBankAccountModal);
}
