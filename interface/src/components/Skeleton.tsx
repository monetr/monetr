import React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={ mergeTailwind('animate-pulse rounded-md bg-muted', className) }
      { ...props }
    />
  );
}
 
export { Skeleton };
