import type React from 'react';

import styles from './Hr.module.scss';

export type HrProps = React.ComponentPropsWithoutRef<'hr'>;

export default function Hr({ className, ...props }: HrProps) {
  return <hr className={className || styles.hr} {...props} />;
}
