import type React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={mergeTailwind('animate-pulse rounded-lg bg-dark-monetr-background-emphasis px-3 py-1.5', className)}
      {...props}
    >
      &nbsp;
    </div>
  );
}

export { Skeleton };
