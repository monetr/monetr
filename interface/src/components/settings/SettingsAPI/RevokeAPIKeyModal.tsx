import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Trash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import Modal, {
  ModalActions,
  ModalContent,
  ModalDescription,
  type ModalRef,
  ModalTitle,
} from '@monetr/interface/components/Modal';
import Typography from '@monetr/interface/components/Typography';
import type ApiKey from '@monetr/interface/models/ApiKey';

export interface RevokeAPIKeyModalProps {
  apiKey: ApiKey;
}

function RevokeAPIKeyModal(props: RevokeAPIKeyModalProps): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent>
        <div>
          <ModalTitle>Revoke API Key?</ModalTitle>
          <ModalDescription>Any automation or script using this key will stop working immediately.</ModalDescription>
        </div>
        <Card>
          <Typography weight='bold'>{props.apiKey.name}</Typography>
          <Typography component='code' ellipsis>
            {props.apiKey.apiKeyId}
          </Typography>
          <Typography component='p' ellipsis size='sm'>
            Created By: <b>[PLACEHOLDER TODO]</b>
          </Typography>
        </Card>
        <ModalActions>
          <Button onClick={modal.remove} variant='secondary'>
            Cancel
          </Button>
          <Button variant='destructive'>
            <Trash />
            Revoke
          </Button>
        </ModalActions>
      </ModalContent>
    </Modal>
  );
}

const revokeApiKeyModal = NiceModal.create<RevokeAPIKeyModalProps>(RevokeAPIKeyModal);

export default revokeApiKeyModal;

export function showRevokeAPIKeyModal(props: RevokeAPIKeyModalProps): Promise<void> {
  return NiceModal.show(revokeApiKeyModal, props);
}

export function closeRevokeAPIKeyModal() {
  return NiceModal.remove(revokeApiKeyModal);
}
