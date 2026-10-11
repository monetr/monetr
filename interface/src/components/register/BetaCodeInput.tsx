import FormTextField from '@monetr/interface/components/FormTextField';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import type { RegisterValues } from '@monetr/interface/pages/register';

import styles from './BetaCodeInput.module.scss';

export default function BetaCodeInput(): React.JSX.Element | null {
  const { data: config } = useAppConfiguration();
  if (!config?.requireBetaCode) {
    return null;
  }

  return (
    <FormTextField<RegisterValues>
      className={styles.input}
      label='Beta Code'
      name='betaCode'
      required
      type='text'
      uppercasetext
    />
  );
}
