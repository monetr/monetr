import { cva, type VariantProps } from 'class-variance-authority';

import Typography, { type TypographyProps } from '@monetr/interface/components/Typography';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Badge.module.scss';

const badgeVariants = cva([styles.badge], {
  variants: {
    variant: {
      brand: styles.brand,
      success: styles.success,
      warning: styles.warning,
      destructive: styles.destructive,
      info: styles.info,
      positive: styles.positive,
    },
  },
  defaultVariants: {
    variant: 'brand',
  },
});

export interface BadgeProps extends Omit<TypographyProps, 'color'>, VariantProps<typeof badgeVariants> {}

export default function Badge({ variant, className, ...props }: BadgeProps): JSX.Element {
  return <Typography {...props} className={mergeClasses(badgeVariants({ variant }), className)} />;
}
