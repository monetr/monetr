import { useCallback, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { FormikHelpers } from 'formik';
import { Trash } from 'lucide-react';
import { useSnackbar } from 'notistack';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import { Switch } from '@monetr/interface/components/Switch';
import Typography from '@monetr/interface/components/Typography';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useRemoveTransaction } from '@monetr/interface/hooks/useRemoveTransaction';
import type Transaction from '@monetr/interface/models/Transaction';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

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

  const submit = useCallback(
    async (
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
        .then(
          () =>
            void enqueueSnackbar('Transaction removed successfully', {
              variant: 'success',
              disableWindowBlurListener: true,
            }),
        )
        .then(() => modal.resolve())
        .then(() => modal.remove())
        .catch(
          (error: ApiError<APIError>) =>
            void enqueueSnackbar(error.response.data.error, {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [enqueueSnackbar, modal, removeTransaction, transaction],
  );

  return (
    <MModal className='md:max-w-md' open={modal.visible} ref={ref}>
      <MForm
        className='h-full flex flex-col gap-stack p-2 justify-between'
        initialValues={initialValues}
        onSubmit={submit}
      >
        {({ setFieldValue, values, isSubmitting }) => (
          <div className='flex flex-col gap-stack'>
              <Typography className='mb-2' size='xl' weight='bold'>
                <Trash />
                Remove Transaction
              </Typography>
              <div className='flex flex-col gap-stack'>
                <Typography size='inherit'>Are you sure you want to remove this transaction?</Typography>
                <ul>
                  <SimilarTransactionItem disableNavigate transactionId={transaction.transactionId} />
                </ul>
                <Typography size='inherit'>You will not be able to undo this action.</Typography>
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
                    checked={values.softDelete}
                    onCheckedChange={() => setFieldValue('softDelete', !values.softDelete)}
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
                    checked={values.adjustsBalance}
                    onCheckedChange={() => setFieldValue('adjustsBalance', !values.adjustsBalance)}
                  />
                </div>
              </div>
              <div className='flex justify-end gap-stack mt-4'>
                <Button disabled={isSubmitting} onClick={modal.remove} variant='secondary'>
                  Cancel
                </Button>
                <FormButton type='submit' variant='destructive'>
                  Remove
                </FormButton>
              </div>
            </div>
        )}
      </MForm>
    </MModal>
  );
}

const removeTransactionModal = NiceModal.create<RemoveTransactionModalProps>(RemoveTransactionModal);

export default removeTransactionModal;

export function showRemoveTransactionModal(props: RemoveTransactionModalProps): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof removeTransactionModal>, unknown>(removeTransactionModal, props);
}
