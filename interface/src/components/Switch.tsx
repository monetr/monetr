import * as React from 'react';
import * as SwitchPrimitives from '@radix-ui/react-switch';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Switch.module.scss';

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitives.Root>,
  React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitives.Root className={mergeTailwind(styles.switchRoot, 'peer', className)} {...props} ref={ref}>
    <SwitchPrimitives.Thumb className={styles.switchThumb} />
  </SwitchPrimitives.Root>
));
Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };
