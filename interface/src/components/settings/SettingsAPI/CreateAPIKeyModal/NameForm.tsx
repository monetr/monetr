import { useCallback } from 'react';
import type { FormikHelpers } from 'formik';
import { Info, Key } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import { ModalActions, ModalContent, ModalDescription, ModalTitle } from '@monetr/interface/components/Modal';
import {
  type CreateAPIKeyMetadata,
  CreateAPIKeySteps,
  closeCreateAPIKeyModal,
} from '@monetr/interface/components/settings/SettingsAPI/CreateAPIKeyModal';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import useCreateApiKey from '@monetr/interface/hooks/useCreateApiKey';
import { useProofOfWork } from '@monetr/interface/hooks/useProofOfWork';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './NameForm.module.scss';

interface CreateAPIKeyModalValues {
  name: string;
}

export default function CreateAPIKeyModalNameForm(): React.JSX.Element {
  const viewContext = useViewContext<CreateAPIKeySteps, CreateAPIKeyMetadata, unknown>();
  const { data: config } = useAppConfiguration();
  const pow = useProofOfWork('create_api_key', Boolean(config?.proofOfWorkEnabled));
  const createApiKey = useCreateApiKey();
  const { enqueueSnackbar } = useSnackbar();
  const submit = useCallback(
    async (values: CreateAPIKeyModalValues, helper: FormikHelpers<CreateAPIKeyModalValues>): Promise<void> => {
      helper.setSubmitting(true);
      return await pow
        .getSolution()
        .then(solution =>
          createApiKey({
            name: values.name,
            challenge: solution?.challenge,
            nonce: solution?.nonce,
          }),
        )
        .then(result => viewContext.updateMetadata({ result }))
        .then(() => viewContext.goToView(CreateAPIKeySteps.Secret))
        .catch((error: ApiError<APIError>) => {
          enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          });
          // To prevent duplicate challenge attempts and failures, reset the challenge so the user can try again.
          pow.reset();
        });
    },
    [createApiKey, enqueueSnackbar, pow, viewContext],
  );

  const initialValues: CreateAPIKeyModalValues = {
    name: '',
  };

  return (
    <MForm initialValues={initialValues} onSubmit={submit}>
      <ModalContent>
        <div>
          <ModalTitle>
            <Key />
            Create API Key
          </ModalTitle>
          <ModalDescription>You'll see the secret once. Store it somewhere safe before closing.</ModalDescription>
        </div>
        <FormTextField
          autoComplete='off'
          data-1p-ignore
          label='Name'
          maxLength={300}
          minLength={1}
          name='name'
          placeholder='e.g. Assistant Integration...'
          required
        />
        <Card className={styles.notice}>
          <Typography color='subtle' size='sm'>
            <Info />
            This key has full read and write access to your monetr account. Treat it like a password, don't commit it to
            git or share it in screenshots.
          </Typography>
        </Card>
        <ModalActions>
          <Button onClick={closeCreateAPIKeyModal} variant='secondary'>
            Cancel
          </Button>
          <FormButton type='submit' variant='primary'>
            <Key />
            Create
          </FormButton>
        </ModalActions>
      </ModalContent>
    </MForm>
  );
}
