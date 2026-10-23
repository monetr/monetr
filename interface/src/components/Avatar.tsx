import type { HTMLAttributes } from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Avatar.module.scss';

const avatarVariants = cva([styles.avatar], {
  variants: {
    size: {
      default: undefined,
      small: styles.sizeSmall,
      large: styles.sizeLarge,
    },
  },
  defaultVariants: {
    size: 'default',
  },
});

export type AvatarProps = HTMLAttributes<HTMLSpanElement> & VariantProps<typeof avatarVariants>;

function Avatar({ className, size, ...props }: AvatarProps): React.JSX.Element {
  return <span className={mergeClasses(avatarVariants({ size }), className)} {...props} />;
}

export type AvatarFallbackProps = HTMLAttributes<HTMLSpanElement>;

function AvatarFallback({ className, ...props }: AvatarFallbackProps): React.JSX.Element {
  return <span className={mergeClasses(styles.avatarFallback, className)} {...props} />;
}

export { Avatar, AvatarFallback, avatarVariants };
