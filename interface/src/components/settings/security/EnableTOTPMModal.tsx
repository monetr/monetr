import React, { Fragment, useEffect, useRef, useState } from 'react';
import QRCode from 'react-qr-code';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { useQueryClient } from '@tanstack/react-query';
import { FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from '@monetr/interface/components/InputOTP';
import MFormButton from '@monetr/interface/components/MButton';
import MDivider from '@monetr/interface/components/MDivider';
import MForm from '@monetr/interface/components/MForm';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import request from '@monetr/interface/util/request';

import { Shield } from 'lucide-react';

interface EnableTOTPValues {
  totp: string;
}

const initialValues: EnableTOTPValues = {
  totp: '',
};

interface TOTPState {
  uri: string;
  recoveryCodes: Array<string>;
}

function EnableTOTPModal(): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const queryClient = useQueryClient();
  const [totpState, setTotpState] = useState<TOTPState | null>(null);

  // As soon as we load the modal, get the TOTP state from the server.
  // TODO This does not handle loading states.
  useEffect(() => {
    if (totpState === null) {
      request().post('/users/security/totp/setup')
        .then(result => setTotpState(result.data))
        .catch(error => {
          console.error('failed to get totp state', error);
        });
    }

  }, [totpState, setTotpState]);

  async function submit(values: EnableTOTPValues, helpers: FormikHelpers<EnableTOTPValues>) {
    helpers.setSubmitting(true);

    return request().post('/users/security/totp/confirm', {
      totp: values.totp,
    })
      .then(() => queryClient.invalidateQueries(['/users/me']))
      .then(() => enqueueSnackbar('Multifactor authentication enabled.', {
        variant: 'success',
        disableWindowBlurListener: true,
      }))
      .then(() => modal.remove())
      .catch(error => enqueueSnackbar(error?.response?.data?.error || 'Failed to enable TOTP.', {
        variant: 'error',
        disableWindowBlurListener: true,
      }));
  }

  return (
    <MModal open={ modal.visible } ref={ ref } className='sm:max-w-sm' onClose={ () => setTotpState(null) }>
      <MForm
        onSubmit={ submit }
        initialValues={ initialValues }
        className='h-full flex flex-col gap-2 p-2 justify-between'
      >
        { ({ setFieldValue, isSubmitting }) => (
          <Fragment>
            <div className='flex flex-col'>
              <MSpan weight='bold' size='xl' className='mb-2'>
                Enable Multifactor Authentication
              </MSpan>
              <div className='flex flex-col items-center gap-4 py-4'>
                <MSpan size='md'>
                  Scan the following QR code with your preferred authenticator app
                </MSpan>
                <div>
                  <div className='bg-white p-2 max-w-fit rounded-md'>
                    <QRCode value={ totpState?.uri || '' } />
                  </div>
                </div>
                <MSpan size='md'>
                  Or enter the secret manually
                </MSpan>
                <MSpan component='code' size='lg' className='py-2 px-4'>
                  { Boolean(totpState?.uri) && (new URL(totpState.uri)).searchParams.get('secret') }
                </MSpan>
                <MDivider className='w-full' />
                <MSpan size='md'>
                  Enter the 6-digit code generated by your app
                </MSpan>
                <InputOTP 
                  name='totp' 
                  maxLength={ 6 } 
                  disabled={ isSubmitting } 
                  onChange={ value => setFieldValue('totp', value) }
                >
                  <InputOTPGroup>
                    <InputOTPSlot index={ 0 } />
                    <InputOTPSlot index={ 1 } />
                    <InputOTPSlot index={ 2 } />
                  </InputOTPGroup>
                  <InputOTPSeparator />
                  <InputOTPGroup>
                    <InputOTPSlot index={ 3 } />
                    <InputOTPSlot index={ 4 } />
                    <InputOTPSlot index={ 5 } />
                  </InputOTPGroup>
                </InputOTP>
              </div>
            </div>
            <div className='flex justify-end gap-2'>
              <MFormButton color='cancel' onClick={ modal.remove } data-testid='close-change-password-modal'>
                Cancel
              </MFormButton>
              <MFormButton color='primary' type='submit'>
                <Shield className='mr-1' />
                Enable TOTP
              </MFormButton>
            </div>
          </Fragment>
        ) }
      </MForm>
    </MModal>
  );
}

const enableTOTPModal = NiceModal.create(EnableTOTPModal);

export default enableTOTPModal;

export function showEnableTOTPModal(): Promise<void> {
  return NiceModal.show<void, React.ComponentProps<typeof enableTOTPModal>, {}>(enableTOTPModal);
}