
import ReCAPTCHA from 'react-google-recaptcha';
import { useFormikContext } from 'formik';
import { LoaderCircle } from 'lucide-react';

import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MCaptchaProps {
  name?: string;
  show?: boolean;
  className?: string;
  'data-testid'?: string;
}

export default function MCaptcha(props: MCaptchaProps): JSX.Element {
  const formikContext = useFormikContext();
  const { data: config } = useAppConfiguration();

  if (!props.show || !config?.ReCAPTCHAKey) {
    return null;
  }

  function onVerify(verification: string): void {
    if (!formikContext?.setFieldValue || !props.name) { return; }

    formikContext.setFieldValue(
      props.name, // Name
      verification, // Value
      false, // Should verify.
    );
  }

  const loading = Boolean(formikContext?.isSubmitting);

  const classes = mergeTailwind(['flex', 'items-center', 'justify-center', 'w-full'], props.className);

  return (
    <div className={classes}>
      {!loading && <ReCAPTCHA data-testid={props['data-testid']} sitekey={config.ReCAPTCHAKey} onChange={onVerify} />}
      {loading && <LoaderCircle className='spin' />}
    </div>
  );
}
