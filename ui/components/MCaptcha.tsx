import React from 'react';
import ReCAPTCHA from 'react-google-recaptcha';
import { CircularProgress } from '@mui/material';
import { useFormikContext } from 'formik';

import clsx from 'clsx';
import { useAppConfiguration } from 'hooks/useAppConfiguration';

export interface MCaptchaProps {
  name?: string;
  show?: boolean;
  className?: string;
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

  const classes = clsx([
    'flex',
    'items-center',
    'justify-center',
    'w-full',
  ], props.className);

  return (
    <div className={ classes }>
      {!loading && <ReCAPTCHA
        sitekey={ ReCAPTCHAKey }
        onChange={ onVerify }
      />}
      { loading && <CircularProgress /> }
    </div>
  );

}
