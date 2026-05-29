import * as React from 'react';
import * as SwitchPrimitives from '@radix-ui/react-switch';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Switch.module.scss';

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitives.Root>,
  React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitives.Root className={mergeClasses(styles.switchRoot, className)} {...props} ref={ref}>
    <SwitchPrimitives.Thumb className={styles.switchThumb} />
  </SwitchPrimitives.Root>
));
Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };
