import FormTextField from '@monetr/interface/components/FormTextField';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

import styles from './BetaCodeInput.module.scss';

export default function BetaCodeInput(): JSX.Element {
  const { data: config } = useAppConfiguration();
  if (!config?.requireBetaCode) {
    return null;
  }

  return (
    <FormTextField className={styles.input} label='Beta Code' name='betaCode' required type='text' uppercasetext />
  );
}
