import React from 'react';
import ReCAPTCHA from 'react-google-recaptcha';
import { useFormikContext } from 'formik';

import { useAppConfiguration } from 'hooks/useAppConfiguration';
import { CircularProgress } from '@mui/material';

export interface MCaptchaProps {
  name?: string;
  show?: boolean;
}

export default function MCaptcha(props: MCaptchaProps): JSX.Element {
  const formikContext = useFormikContext();
  const { ReCAPTCHAKey } = useAppConfiguration();

  if (!props.show || !ReCAPTCHAKey) {
    return null;
  }

  function onVerify(verification: string): void {
    if (!formikContext?.setFieldValue || !props.name) return;

    formikContext.setFieldValue(
      props.name, // Name
      verification, // Value
      false, // Should verify.
    );
  }

  const loading = Boolean(formikContext?.isSubmitting);

  return (
    <div className="flex items-center justify-center w-full">
      {!loading && <ReCAPTCHA
        sitekey={ ReCAPTCHAKey }
        onChange={ onVerify }
      />}
      { loading && <CircularProgress /> }
    </div>
  );

}
