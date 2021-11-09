import { useSelector } from 'react-redux';
import { getReCAPTCHAKey } from 'shared/bootstrap/selectors';
import React from 'react';
import { CircularProgress } from '@mui/material';
import ReCAPTCHA from 'react-google-recaptcha';

export interface PropTypes {
  show?: boolean;
  loading?: boolean;
  onVerify: (verification: string) => void;
}

const CaptchaMaybe = (props: PropTypes): JSX.Element => {
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
      { loading && <CircularProgress/> }
    </div>
  );
}

export default CaptchaMaybe;