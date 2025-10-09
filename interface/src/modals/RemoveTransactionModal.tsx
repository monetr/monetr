import React, { Fragment, useCallback, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { AxiosError } from 'axios';
import { FormikHelpers } from 'formik';
import { Trash } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import { Switch } from '@monetr/interface/components/Switch';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useRemoveTransaction } from '@monetr/interface/hooks/useRemoveTransaction';
import Transaction from '@monetr/interface/models/Transaction';
import { APIError } from '@monetr/interface/util/request';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

interface RemoveTransactionModalProps {
  transaction: Transaction;
}

interface RemoveTransactionModalValues {
  adjustsBalance: boolean;
  softDelete: boolean;
}

const initialValues: RemoveTransactionModalValues = {
  adjustsBalance: false,
  softDelete: true,
};

function RemoveTransactionModal(props: RemoveTransactionModalProps): JSX.Element {
  const { transaction } = props;
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const removeTransaction = useRemoveTransaction();

  const submit = useCallback(async (
    values: RemoveTransactionModalValues,
    helpers: FormikHelpers<RemoveTransactionModalValues>,
  ): Promise<void> => {
    helpers.setSubmitting(true);

    // Send the delete request to the server and handle any changes returned.
    return await removeTransaction({
      transaction,
      softDelete: values.softDelete,
      adjustsBalance: values.adjustsBalance,
    })
      .then(() => void enqueueSnackbar('Transaction removed successfully', {
        variant: 'success',
        disableWindowBlurListener: true,
      }))
      .then(() => modal.resolve())
      .then(() => modal.remove())
      .catch((error: AxiosError<APIError>) => void enqueueSnackbar(error.response.data['error'], {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helpers.setSubmitting(false));
  }, [enqueueSnackbar, modal, removeTransaction, transaction]);

  return (
    <MModal open={ modal.visible } ref={ ref } className='md:max-w-md'>
      <MForm
        onSubmit={ submit }
        initialValues={ initialValues }
        className='h-full flex flex-col gap-stack p-2 justify-between'
      >
        { ({ setFieldValue, values, isSubmitting }) => (
          <Fragment>
            <div className='flex flex-col gap-stack'>
              <MSpan weight='bold' size='xl' className='mb-2'>
                <Trash />
                Remove Transaction
              </MSpan>
              <div className='flex flex-col gap-stack'>
                <MSpan>
                  Are you sure you want to remove this transaction?
                </MSpan>
                <ul>
                  <SimilarTransactionItem
                    transactionId={ transaction.transactionId }
                    disableNavigate
                  />
                </ul>
                <MSpan>
                  You will not be able to undo this action.
                </MSpan>
                <div className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string gap-component'>
                  <div className='gap-component'>
                    <label className='text-sm font-medium text-dark-monetr-content-emphasis cursor-pointer'>
                      Prevent Re-Creation
                    </label>
                    <p className='text-sm text-dark-monetr-content'>
                      Prevent this transaction from be re-created if it is present in future file imports?
                    </p>
                  </div>
                  <Switch
                    checked={ values['softDelete'] }
                    onCheckedChange={ () => setFieldValue('softDelete', !values['softDelete']) }
                  />
                </div>
                <div className='flex flex-row items-center justify-between rounded-lg ring-1 p-2 ring-dark-monetr-border-string gap-component'>
                  <div className='gap-component'>
                    <label className='text-sm font-medium text-dark-monetr-content-emphasis cursor-pointer'>
                      Adjust Balance
                    </label>
                    <p className='text-sm text-dark-monetr-content'>
                      Update your account balance as if this transaction was reversed?
                    </p>
                  </div>
                  <Switch
                    checked={ values['adjustsBalance'] }
                    onCheckedChange={ () => setFieldValue('adjustsBalance', !values['adjustsBalance']) }
                  />
                </div>
              </div>
              <div className='flex justify-end gap-stack mt-4'>
                <Button disabled={ isSubmitting } variant='secondary' onClick={ modal.remove }>
                  Cancel
                </Button>
                <FormButton variant='destructive' type='submit'>
                  Remove
                </FormButton>
              </div>
            </div>
          </Fragment>
        ) }
      </MForm>
    </MModal>
  );
}

const removeTransactionModal = NiceModal.create<RemoveTransactionModalProps>(RemoveTransactionModal);

export default removeTransactionModal;

export function showRemoveTransactionModal(props: RemoveTransactionModalProps): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof removeTransactionModal>, {}>(removeTransactionModal, props);
}

