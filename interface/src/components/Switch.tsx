import * as React from 'react';
import * as SwitchPrimitives from '@radix-ui/react-switch';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitives.Root>,
  React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitives.Root
    className={mergeTailwind(
      'peer inline-flex items-center h-6 w-11 shrink-0',
      'cursor-pointer',
      'border-2 border-transparent rounded-full',
      'transition-colors',
      // Focuses
      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
      'focus-visible:ring-offset-2 focus-visible:ring-offset-background',
      // Disabled
      'disabled:cursor-not-allowed disabled:opacity-50',
      // Checked
      'data-[state=checked]:bg-dark-monetr-green data-[state=unchecked]:bg-dark-monetr-background-subtle',
      className,
    )}
    {...props}
    ref={ref}
  >
    <SwitchPrimitives.Thumb
      className={mergeTailwind(
        'pointer-events-none',
        'block h-5 w-5 rounded-full',
        'bg-dark-monetr-content-emphasis shadow-lg',
        'ring-0 transition-transform',
        'data-[state=checked]:translate-x-5 data-[state=unchecked]:translate-x-0',
      )}
    />
  </SwitchPrimitives.Root>
));
Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };
