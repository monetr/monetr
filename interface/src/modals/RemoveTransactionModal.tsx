import { useCallback, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { FormikHelpers } from 'formik';
import { Trash } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import SwitchCard from '@monetr/interface/components/SwitchCard';
import Typography from '@monetr/interface/components/Typography';
import SimilarTransactionItem from '@monetr/interface/components/transactions/SimilarTransactionItem';
import { useRemoveTransaction } from '@monetr/interface/hooks/useRemoveTransaction';
import type Transaction from '@monetr/interface/models/Transaction';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';
import { useSnackbar } from '@monetr/notify';

import styles from './RemoveTransactionModal.module.scss';

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
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
        {({ setFieldValue, values, isSubmitting }) => (
          <div className={styles.body}>
            <Typography className={styles.heading} size='xl' weight='bold'>
              <Trash />
              Remove Transaction
            </Typography>
            <div className={styles.options}>
              <Typography size='inherit'>Are you sure you want to remove this transaction?</Typography>
              <ul>
                <SimilarTransactionItem disableNavigate transactionId={transaction.transactionId} />
              </ul>
              <Typography size='inherit'>You will not be able to undo this action.</Typography>
              <SwitchCard
                checked={values.softDelete}
                description='Prevent this transaction from be re-created if it is present in future file imports?'
                label='Prevent Re-Creation'
                onCheckedChange={() => setFieldValue('softDelete', !values.softDelete)}
              />
              <SwitchCard
                checked={values.adjustsBalance}
                description='Update your account balance as if this transaction was reversed?'
                label='Adjust Balacne'
                onCheckedChange={() => setFieldValue('adjustsBalance', !values.adjustsBalance)}
              />
            </div>
            <div className={styles.actions}>
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
  return NiceModal.show<
    void,
    ExtractProps<typeof removeTransactionModal>,
    Partial<ExtractProps<typeof removeTransactionModal>>
  >(removeTransactionModal, props);
}
