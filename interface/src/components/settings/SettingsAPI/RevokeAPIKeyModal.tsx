import { useCallback, useRef, useState } from 'react';
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
import APIKeyItem from '@monetr/interface/components/settings/SettingsAPI/APIKeyItem';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useProofOfWork } from '@monetr/interface/hooks/useProofOfWork';
import useRemoveApiKey from '@monetr/interface/hooks/useRemoveApiKey';
import type ApiKey from '@monetr/interface/models/ApiKey';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export interface RevokeAPIKeyModalProps {
  apiKey: ApiKey;
}

function RevokeAPIKeyModal(props: RevokeAPIKeyModalProps): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const { data: config } = useAppConfiguration();
  const pow = useProofOfWork('delete_api_key', Boolean(config?.proofOfWorkEnabled));
  const removeApiKey = useRemoveApiKey();
  const [submitting, setSubmitting] = useState(false);

  const submit = useCallback(async () => {
    setSubmitting(true);
    return await pow
      .getSolution()
      .then(solution =>
        removeApiKey({
          apiKeyId: props.apiKey.apiKeyId,
          challenge: solution?.challenge,
          nonce: solution?.nonce,
        }),
      )
      .then(() => modal.remove())
      .catch((error: ApiError<APIError>) => {
        setSubmitting(false);
        enqueueSnackbar(error.response.data.error, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
        pow.reset();
      });
  }, [removeApiKey, enqueueSnackbar, modal.remove, pow.getSolution, pow.reset, props.apiKey.apiKeyId]);

  return (
    <Modal open={modal.visible} ref={ref}>
      <ModalContent>
        <div>
          <ModalTitle>Revoke API Key?</ModalTitle>
          <ModalDescription>Any automation or script using this key will stop working immediately.</ModalDescription>
        </div>
        <APIKeyItem apiKey={props.apiKey} hideRevoke />
        <ModalActions>
          <Button disabled={submitting} onClick={modal.remove} variant='secondary'>
            Cancel
          </Button>
          <Button disabled={submitting} onClick={submit} variant='destructive'>
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
