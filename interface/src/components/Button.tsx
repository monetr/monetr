import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Button.module.css';

const buttonVariants = cva([styles.button], {
  variants: {
    variant: {
      primary: styles.primary,
      secondary: styles.secondary,
      outlined: styles.outlinend,
      destructive: styles.destructive,
      text: styles.text,
    },
    size: {
      default: 'h-8 text-sm',
      select: 'h-[38px]',
      md: 'min-h-10',
    },
  },
  defaultVariants: {
    size: 'default',
    variant: 'primary',
  },
});

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant: color, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button';
    return (
      <Comp
        className={mergeTailwind(buttonVariants({ variant: color, size, className }))}
        ref={ref}
        tabIndex={0}
        type='button'
        {...props}
      />
    );
  },
);
Button.displayName = 'Button';

export { Button, buttonVariants };
