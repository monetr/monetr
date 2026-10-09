import { useRef, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Archive } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Modal, {
  ModalActions,
  ModalContent,
  ModalDescription,
  type ModalRef,
  ModalTitle,
} from '@monetr/interface/components/Modal';
import { useArchiveBankAccount } from '@monetr/interface/hooks/useArchiveBankAccount';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export interface ArchiveBankAccountModalProps {
  bankAccount: BankAccount;
}

function ArchiveBankAccountModal(props: ArchiveBankAccountModalProps): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const archiveBankAccount = useArchiveBankAccount();
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    setSubmitting(true);
    return await archiveBankAccount(props.bankAccount.bankAccountId)
      .then(() => modal.resolve())
      .then(() => modal.remove())
      .catch((error: ApiError<APIError>) => {
        setSubmitting(false);
        enqueueSnackbar(error?.response?.data?.error || 'Failed to archive bank account', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent data-testid='archive-bank-account-modal'>
        <div>
          <ModalTitle>Archive Bank Account?</ModalTitle>
          <ModalDescription>Are you sure you want to archive {props.bankAccount.name}?</ModalDescription>
        </div>
        <ModalActions>
          <Button
            data-testid='close-archive-bank-account-modal'
            disabled={submitting}
            onClick={modal.remove}
            variant='secondary'
          >
            Cancel
          </Button>
          <Button
            data-testid='archive-bank-account-confirm'
            disabled={submitting}
            onClick={submit}
            variant='destructive'
          >
            <Archive />
            Archive
          </Button>
        </ModalActions>
      </ModalContent>
    </Modal>
  );
}

const archiveBankAccountModal = NiceModal.create<ArchiveBankAccountModalProps>(ArchiveBankAccountModal);

export default archiveBankAccountModal;

export function showArchiveBankAccountModal(props: ArchiveBankAccountModalProps): Promise<void> {
  return NiceModal.show(archiveBankAccountModal, props);
}
