import { useId } from 'react';
import type { SwitchProps } from '@radix-ui/react-switch';

import { Switch } from '@monetr/interface/components/Switch';

import styles from './SwitchCard.module.scss';

interface SwitchCardProps extends Omit<SwitchProps, 'id'> {
  label: string;
  description: string;
}

export default function SwitchCard(props: SwitchCardProps): React.JSX.Element {
  const { label, description, ...switchProps } = props;
  const id = useId();
  return (
    <div className={styles.optionRow}>
      <div className={styles.optionText}>
        <label className={styles.optionLabel} htmlFor={id}>
          {label}
        </label>
        <p className={styles.optionDescription}>{description}</p>
      </div>
      <Switch id={id} {...switchProps} />
    </div>
  );
}
