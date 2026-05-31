import { useFormikContext } from 'formik';
import { LoaderCircle } from 'lucide-react';
import ReCAPTCHA from 'react-google-recaptcha';

import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './MCaptcha.module.scss';

export interface MCaptchaProps {
  name?: string;
  show?: boolean;
  className?: string;
  'data-testid'?: string;
}

export default function MCaptcha(props: MCaptchaProps): React.ReactNode {
  const formikContext = useFormikContext();
  const { data: config } = useAppConfiguration();

  if (!props.show || !config?.ReCAPTCHAKey) {
    return null;
  }

  function onVerify(verification: string | null): void {
    if (!formikContext?.setFieldValue || !props.name) {
      return;
    }

    formikContext.setFieldValue(
      props.name, // Name
      verification, // Value
      false, // Should verify.
    );
  }

  const loading = Boolean(formikContext?.isSubmitting);

  const classes = mergeClasses(styles.root, props.className);

  return (
    <div className={classes}>
      {!loading && <ReCAPTCHA data-testid={props['data-testid']} onChange={onVerify} sitekey={config.ReCAPTCHAKey} />}
      {loading && <LoaderCircle className='spin' />}
    </div>
  );
}
