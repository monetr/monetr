import { useRef, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Trash } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Modal, {
  ModalActions,
  ModalContent,
  ModalDescription,
  type ModalRef,
  ModalTitle,
} from '@monetr/interface/components/Modal';
import { useRemoveSpending } from '@monetr/interface/hooks/useRemoveSpending';
import type Spending from '@monetr/interface/models/Spending';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export interface RemoveExpenseModalProps {
  spending: Spending;
}

function RemoveExpenseModal(props: RemoveExpenseModalProps): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const removeSpending = useRemoveSpending();
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    setSubmitting(true);
    return await removeSpending(props.spending.spendingId)
      .then(() => modal.resolve())
      .then(() => modal.remove())
      .catch((error: ApiError<APIError>) => {
        setSubmitting(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to remove expense', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent data-testid='remove-expense-modal'>
        <div>
          <ModalTitle>Remove Expense?</ModalTitle>
          <ModalDescription>Are you sure you want to remove {props.spending.name}?</ModalDescription>
        </div>
        <ModalActions>
          <Button
            data-testid='close-remove-expense-modal'
            disabled={submitting}
            onClick={modal.remove}
            variant='secondary'
          >
            Cancel
          </Button>
          <Button data-testid='remove-expense-confirm' disabled={submitting} onClick={submit} variant='destructive'>
            <Trash />
            Remove
          </Button>
        </ModalActions>
      </ModalContent>
    </Modal>
  );
}

const removeExpenseModal = NiceModal.create<RemoveExpenseModalProps>(RemoveExpenseModal);

export default removeExpenseModal;

export function showRemoveExpenseModal(props: RemoveExpenseModalProps): Promise<void> {
  return NiceModal.show(removeExpenseModal, props);
}
