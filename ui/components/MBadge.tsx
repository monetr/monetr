import React from 'react';

import MSpan from './MSpan';
import { ReactElement } from './types';

import mergeTailwind from 'util/mergeTailwind';


export interface MBadgeProps {
  className?: string;
  children: ReactElement;
}

export default function MBadge(props: MBadgeProps): JSX.Element {

  const classes = mergeTailwind(
    'bg-monetr-brand',
    'dark:text-dark-monetr-content-emphasis',
    'px-1.5',
    'py-0.5',
    'rounded-md',
    'text-sm',
    props.className,
  );

  return (
    <MSpan className={ classes }>
      { props.children }
    </MSpan>
  );
}
