import React from 'react';
import ReCAPTCHA from 'react-google-recaptcha';
import { useSelector } from 'react-redux';
import { CircularProgress } from '@mui/material';

import { getReCAPTCHAKey } from 'shared/bootstrap/selectors';

export interface PropTypes {
  show?: boolean;
  loading?: boolean;
  onVerify: (verification: string) => void;
}

export default function CaptchaMaybe(props: PropTypes): JSX.Element {
  const reCaptchaKey = useSelector(getReCAPTCHAKey);

  const { show, loading, onVerify } = props;

  if (!show || !reCaptchaKey) {
    return null;
  }

  return (
    <div className="flex items-center justify-center w-full">
      { !loading && <ReCAPTCHA
        sitekey={ reCaptchaKey }
        onChange={ onVerify }
      /> }
      { loading && <CircularProgress /> }
    </div>
  );
}
