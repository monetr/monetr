import type React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Skeleton.module.scss';

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={mergeTailwind(styles.skeleton, className)} {...props}>
      &nbsp;
    </div>
  );
}

export { Skeleton };
