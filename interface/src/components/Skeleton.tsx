import type React from 'react';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Skeleton.module.scss';

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={mergeClasses(styles.skeleton, className)} {...props}>
      &nbsp;
    </div>
  );
}

export { Skeleton };
