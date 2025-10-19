import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

const buttonVariants = cva(
  [
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus:outline-none',
    'font-semibold',
    'inline-flex items-center gap-1 justify-center',
    'rounded-lg',
    'enabled:active:brightness-110',
    '[&_svg]:pointer-events-none [&_svg]:size-4',
    'disabled:pointer-events-none',
  ],
  {
    variants: {
      variant: {
        primary: [
          'bg-monetr-brand enabled:hover:bg-dark-monetr-brand-subtle disabled:bg-purple-200',
          'focus-visible:outline-purple-600',
          'text-dark-monetr-content-emphasis',
          'shadow-sm',
        ],
        secondary: [
          'bg-dark-monetr-background-subtle enabled:hover:bg-dark-monetr-background-emphasis',
          'ring-1 ring-inset ring-dark-monetr-border disabled:ring-dark-monetr-border-subtle',
          'text-dark-monetr-content-emphasis disabled:text-dark-monetr-content-subtle',
          'focus-visible:outline-purple-200',
        ],
        outlined: [
          'ring-1 ring-inset enabled:focus:ring-2 enabled:hover:ring-2',
          'enabled:hover:ring-dark-monetr-brand enabled:focus:ring-dark-monetr-brand',
          'ring-dark-monetr-border-string disabled:ring-dark-monetr-border-subtle',
          'text-dark-monetr-content-emphasis disabled:text-dark-monetr-content-muted',
        ],
        destructive: [
          'bg-red-600 enabled:hover:bg-red-500 disabled:bg-red-200',
          'focus-visible:outline-red-600',
          'text-dark-monetr-content-emphasis',
          'shadow-sm',
        ],
        text: [
          'enabled:hover:bg-dark-monetr-background-emphasis',
          'text-dark-monetr-content-emphasis disabled:text-dark-monetr-content-muted',
          'focus-visible:outline-purple-200',
        ],
      },
      size: {
        default: 'h-8 text-sm px-3 py-1.5',
        select: 'h-[38px] px-3 py-1.5',
        md: 'min-h-10 px-3 py-1.5',
      },
    },
    defaultVariants: {
      size: 'default',
      variant: 'primary',
    },
  },
);

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
