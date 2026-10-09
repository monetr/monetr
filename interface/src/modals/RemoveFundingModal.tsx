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
import { useRemoveFundingSchedule } from '@monetr/interface/hooks/useRemoveFundingSchedule';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export interface RemoveFundingModalProps {
  funding: FundingSchedule;
}

function RemoveFundingModal(props: RemoveFundingModalProps): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const removeFundingSchedule = useRemoveFundingSchedule();
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    setSubmitting(true);
    return await removeFundingSchedule(props.funding)
      .then(() => modal.resolve())
      .then(() => modal.remove())
      .catch((error: ApiError<APIError>) => {
        setSubmitting(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to remove funding schedule', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent data-testid='remove-funding-modal'>
        <div>
          <ModalTitle>Remove Funding Schedule?</ModalTitle>
          <ModalDescription>Are you sure you want to remove {props.funding.name}?</ModalDescription>
        </div>
        <ModalActions>
          <Button
            data-testid='close-remove-funding-modal'
            disabled={submitting}
            onClick={modal.remove}
            variant='secondary'
          >
            Cancel
          </Button>
          <Button data-testid='remove-funding-confirm' disabled={submitting} onClick={submit} variant='destructive'>
            <Trash />
            Remove
          </Button>
        </ModalActions>
      </ModalContent>
    </Modal>
  );
}

const removeFundingModal = NiceModal.create<RemoveFundingModalProps>(RemoveFundingModal);

export default removeFundingModal;

export function showRemoveFundingModal(props: RemoveFundingModalProps): Promise<void> {
  return NiceModal.show(removeFundingModal, props);
}
