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

export interface RemoveGoalModalProps {
  spending: Spending;
}

function RemoveGoalModal(props: RemoveGoalModalProps): React.JSX.Element {
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
        enqueueSnackbar(error?.response?.data?.error || 'Failed to remove goal', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent data-testid='remove-goal-modal'>
        <div>
          <ModalTitle>Remove Goal?</ModalTitle>
          <ModalDescription>Are you sure you want to remove {props.spending.name}?</ModalDescription>
        </div>
        <ModalActions>
          <Button
            data-testid='close-remove-goal-modal'
            disabled={submitting}
            onClick={modal.remove}
            variant='secondary'
          >
            Cancel
          </Button>
          <Button data-testid='remove-goal-confirm' disabled={submitting} onClick={submit} variant='destructive'>
            <Trash />
            Remove
          </Button>
        </ModalActions>
      </ModalContent>
    </Modal>
  );
}

const removeGoalModal = NiceModal.create<RemoveGoalModalProps>(RemoveGoalModal);

export default removeGoalModal;

export function showRemoveGoalModal(props: RemoveGoalModalProps): Promise<void> {
  return NiceModal.show(removeGoalModal, props);
}
