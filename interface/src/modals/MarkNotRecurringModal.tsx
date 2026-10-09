import { useRef, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { RepeatOff } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Modal, {
  ModalActions,
  ModalContent,
  ModalDescription,
  type ModalRef,
  ModalTitle,
} from '@monetr/interface/components/Modal';
import { useRemoveTransactionRecurring } from '@monetr/interface/hooks/useRemoveTransactionRecurring';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export interface MarkNotRecurringModalProps {
  recurring: TransactionRecurring;
}

function MarkNotRecurringModal(props: MarkNotRecurringModalProps): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const removeTransactionRecurring = useRemoveTransactionRecurring();
  // The recurring transaction doesn't have a name of its own, it uses the name of its similar transactions group.
  const { data: cluster } = useTransactionCluster(props.recurring.transactionClusterId);
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    setSubmitting(true);
    return await removeTransactionRecurring({
      transactionRecurringId: props.recurring.transactionRecurringId,
      bankAccountId: props.recurring.bankAccountId,
    })
      .then(() => modal.resolve())
      .then(() => modal.remove())
      .catch((error: ApiError<APIError>) => {
        setSubmitting(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to mark as not recurring', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent data-testid='mark-not-recurring-modal'>
        <div>
          <ModalTitle>Mark As Not Recurring?</ModalTitle>
          <ModalDescription>
            Are you sure {cluster?.name ?? 'this'} isn&apos;t recurring? monetr will stop suggesting it, and you
            won&apos;t be able to unmark this later.
          </ModalDescription>
        </div>
        <ModalActions>
          <Button
            data-testid='close-mark-not-recurring-modal'
            disabled={submitting}
            onClick={modal.remove}
            variant='secondary'
          >
            Cancel
          </Button>
          <Button data-testid='mark-not-recurring-confirm' disabled={submitting} onClick={submit} variant='destructive'>
            <RepeatOff />
            Not Recurring
          </Button>
        </ModalActions>
      </ModalContent>
    </Modal>
  );
}

const markNotRecurringModal = NiceModal.create<MarkNotRecurringModalProps>(MarkNotRecurringModal);

export default markNotRecurringModal;

export function showMarkNotRecurringModal(props: MarkNotRecurringModalProps): Promise<void> {
  return NiceModal.show(markNotRecurringModal, props);
}
