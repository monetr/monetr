import type React from 'react';
import { cva } from 'class-variance-authority';

import { textSizes, textWeights } from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './TextLink.module.scss';

import { Link } from 'wouter';

export const textLinkVariants = cva([styles.textLink], {
  variants: {
    variant: {
      primary: styles.primary,
      secondary: styles.secondary,
    },
    size: textSizes,
    weight: textWeights,
  },
  defaultVariants: {
    variant: 'primary',
    size: 'md',
    weight: 'semibold',
  },
});

type VariantProps = Omit<Parameters<typeof textLinkVariants>[0], 'className' | 'class'>;

export type TextLinkProps = VariantProps &
  Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> & {
    to: string;
  };

export default function TextLink({ variant, size, className, ...props }: TextLinkProps): React.JSX.Element {
  return (
    <Link
      className={mergeTailwind(
        textLinkVariants({
          variant,
          size,
        }),
        className,
      )}
      tabIndex={0}
      {...props}
    />
  );
}
