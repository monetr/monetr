import type React from 'react';

import styles from './Typography.module.scss';

export type TypographyVariant = 'body' | 'footer';

export type TypographyProps = React.ComponentPropsWithoutRef<'p'> & {
  variant?: TypographyVariant;
};

const variantStyles: Record<TypographyVariant, string> = {
  body: styles.body,
  footer: styles.footer,
};

export function Typography({ variant = 'body', className, ...props }: TypographyProps) {
  return <p className={className || variantStyles[variant]} {...props} />;
}
