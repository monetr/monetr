import React from 'react';
import ReCAPTCHA from 'react-google-recaptcha';
import { CircularProgress } from '@mui/material';

import { useAppConfiguration } from 'hooks/useAppConfiguration';

export interface PropTypes {
  show?: boolean;
  loading?: boolean;
  onVerify: (verification: string) => void;
}

export default function CaptchaMaybe(props: PropTypes): JSX.Element {
  const {
    ReCAPTCHAKey,
  } = useAppConfiguration();

  const { show, loading, onVerify } = props;

  if (!show || !ReCAPTCHAKey) {
    return null;
  }

  return (
    <div className="flex items-center justify-center w-full">
      { !loading && <ReCAPTCHA
        sitekey={ ReCAPTCHAKey }
        onChange={ onVerify }
      /> }
      { loading && <CircularProgress /> }
    </div>
  );
}
