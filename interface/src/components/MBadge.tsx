import React from 'react';

import MSpan, { type MSpanProps } from './MSpan';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MBadgeProps extends Omit<MSpanProps, 'color'> {}

export default function MBadge(props: MBadgeProps): JSX.Element {
  const classes = mergeTailwind(
    'bg-monetr-brand',
    'dark:text-dark-monetr-content-emphasis',
    'px-2',
    'py-0.5',
    'rounded-lg',
    props.className,
  );

  return <MSpan {...props} className={classes} />;
}
