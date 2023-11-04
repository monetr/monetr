import React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MDividerProps {
  className?: string;
}

export default function MDivider(props: MDividerProps): JSX.Element {
  const className = mergeTailwind(
    'border-0 border-b-[thin] dark:border-dark-monetr-border',
    props.className,
  );

  return (
    <hr className={ className } />
  );
}
