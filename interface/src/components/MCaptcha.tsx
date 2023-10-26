import React from 'react';
import ReCAPTCHA from 'react-google-recaptcha';
import { CircularProgress } from '@mui/material';
import { useFormikContext } from 'formik';

import { useAppConfiguration } from 'hooks/useAppConfiguration';
import mergeTailwind from 'util/mergeTailwind';

export interface MCaptchaProps {
  name?: string;
  show?: boolean;
  className?: string;
  ['data-testid']?: string;
}

export default function MCaptcha(props: MCaptchaProps): JSX.Element {
  const formikContext = useFormikContext();
  const config = useAppConfiguration();

  if (!props.show || !config?.ReCAPTCHAKey) {
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

  const classes = mergeTailwind([
    'flex',
    'items-center',
    'justify-center',
    'w-full',
  ], props.className);

  return (
    <div className={ classes }>
      {!loading && <ReCAPTCHA
        data-testid={ props['data-testid'] }
        sitekey={ config.ReCAPTCHAKey }
        onChange={ onVerify }
      />}
      {loading && <CircularProgress />}
    </div>
  );

}
