import React from 'react';

import MSpan, { MSpanProps } from './MSpan';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';


export interface MBadgeProps extends Omit<MSpanProps, 'color'>{
}

export default function MBadge(props: MBadgeProps): JSX.Element {

  const classes = mergeTailwind(
    'bg-monetr-brand',
    'dark:text-dark-monetr-content-emphasis',
    'px-1.5',
    'py-0.5',
    'rounded-md',
    props.className,
  );

  return (
    <MSpan { ...props } className={ classes } />
  );
}
