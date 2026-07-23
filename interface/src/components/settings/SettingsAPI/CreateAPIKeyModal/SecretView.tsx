import { Key, KeyRound, SquareAsterisk } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Code from '@monetr/interface/components/Code';
import { ModalActions, ModalContent, ModalDescription, ModalTitle } from '@monetr/interface/components/Modal';
import {
  type CreateAPIKeyMetadata,
  type CreateAPIKeySteps,
  closeCreateAPIKeyModal,
} from '@monetr/interface/components/settings/SettingsAPI/CreateAPIKeyModal';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';

export default function CreateAPIKeyModalSecretView(): React.JSX.Element {
  const viewContext = useViewContext<CreateAPIKeySteps, CreateAPIKeyMetadata, unknown>();

  return (
    <ModalContent>
      <div>
        <ModalTitle>
          <Key />
          API Key Created
        </ModalTitle>
        <ModalDescription>Copy it now, we can't show this secret again.</ModalDescription>
      </div>
      <Typography>{viewContext.metadata.result?.name}</Typography>
      <Code copy={viewContext.metadata.result?.apiKeyId} icon={KeyRound} label='Key ID'>
        {viewContext.metadata.result?.apiKeyId}
      </Code>
      <Code copy={viewContext.metadata.result?.secret} icon={SquareAsterisk} label='Secret'>
        {viewContext.metadata.result?.secret}
      </Code>
      <Typography color='subtle' component='p' size='sm'>
        monetr authenticates with HTTP basic authentication, send the <b>Key ID</b> as the username and the{' '}
        <b>secret</b> as the password.
      </Typography>
      <ModalActions>
        <Button onClick={closeCreateAPIKeyModal} variant='primary'>
          Done
        </Button>
      </ModalActions>
    </ModalContent>
  );
}
