import type React from 'react';
import styles from './Text.module.scss';

export type TextVariant = 'body' | 'footer';

export type TextProps = React.ComponentPropsWithoutRef<'p'> & {
  variant?: TextVariant;
};

const variantStyles: Record<TextVariant, string> = {
  body: styles.body,
  footer: styles.footer,
};

export function Text({ variant = 'body', className, ...props }: TextProps) {
  return <p className={className || variantStyles[variant]} {...props} />;
}
