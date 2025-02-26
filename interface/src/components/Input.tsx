import * as React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={mergeTailwind(
          'flex h-10 w-full rounded-md border border-dark-monetr-border bg-transparent px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-dark-monetr-content-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-monetr-brand focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
Input.displayName = 'Input';

export { Input };
