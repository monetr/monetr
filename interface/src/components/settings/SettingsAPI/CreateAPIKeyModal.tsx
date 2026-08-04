import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';

import Modal, { type ModalRef } from '@monetr/interface/components/Modal';
import NameForm from '@monetr/interface/components/settings/SettingsAPI/CreateAPIKeyModal/NameForm';
import SecretView from '@monetr/interface/components/settings/SettingsAPI/CreateAPIKeyModal/SecretView';
import { ViewManager } from '@monetr/interface/components/ViewManager';
import type { CreateApiKeyResponse } from '@monetr/interface/hooks/useCreateApiKey';

export interface CreateAPIKeyMetadata {
  result?: CreateApiKeyResponse;
}

export enum CreateAPIKeySteps {
  Name = 'CreateAPIKeySteps.Name',
  Secret = 'CreateAPIKeySteps.Secret',
}

function CreateAPIKeyModal(): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<ModalRef>(null);

  return (
    <Modal open={modal.visible} ref={ref}>
      <ViewManager<CreateAPIKeySteps, CreateAPIKeyMetadata, unknown>
        initialView={CreateAPIKeySteps.Name}
        viewComponents={{
          [CreateAPIKeySteps.Name]: NameForm,
          [CreateAPIKeySteps.Secret]: SecretView,
        }}
      />
    </Modal>
  );
}

const createApiKeyModal = NiceModal.create(CreateAPIKeyModal);

export default createApiKeyModal;

export function showCreateAPIKeyModal(): Promise<void> {
  return NiceModal.show(createApiKeyModal);
}

export function closeCreateAPIKeyModal() {
  return NiceModal.remove(createApiKeyModal);
}
