import type { ButtonHTMLAttributes } from 'react';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Switch.module.scss';

export interface SwitchProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}

export function Switch({ className, checked, onCheckedChange, ...props }: SwitchProps): React.JSX.Element {
  const state = checked ? 'checked' : 'unchecked';

  function toggle() {
    onCheckedChange(!checked);
  }

  return (
    <button
      {...props}
      aria-checked={checked}
      className={mergeClasses(styles.switchRoot, className)}
      data-state={state}
      onClick={toggle}
      role='switch'
      type='button'
    >
      <span className={styles.switchThumb} data-state={state} />
    </button>
  );
}
