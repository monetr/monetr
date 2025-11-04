import * as React from 'react';
import type { CheckedState } from '@radix-ui/react-checkbox';
import * as CheckboxPrimitive from '@radix-ui/react-checkbox';
import { Check } from 'lucide-react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Checkbox.module.scss';

const Checkbox = React.forwardRef<
  React.ElementRef<typeof CheckboxPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root>
>(({ className, ...props }, ref) => (
  <CheckboxPrimitive.Root ref={ref} className={mergeTailwind(styles.checkboxRoot, 'peer', className)} {...props}>
    <CheckboxPrimitive.Indicator className={styles.checkboxIndicator}>
      <Check className={styles.checkboxCheck} />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox, type CheckedState };
