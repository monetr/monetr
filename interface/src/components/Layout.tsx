import { cva } from 'class-variance-authority';

import styles from './Layout.module.scss';

export const widths = {
  default: undefined,
  '1/4': styles.widthQuarter,
  '1/3': styles.widthThird,
  '1/2': styles.widthHalf,
  full: styles.widthFull,
  screen: styles.widthScreen,
} as const;

export const maxWidths = {
  default: undefined,
  small: styles.maxWidthSmall,
  medium: styles.maxWidthMedium,
  large: styles.maxWidthLarge,
  extraLarge: styles.maxWidthExtraLarge,
} as const;

export const heights = {
  default: undefined,
  full: styles.heightFull,
  screen: styles.heightScreen,
} as const;

export const layoutVariants = cva([], {
  variants: {
    width: widths,
    height: heights,
    maxWidth: maxWidths,
  },
  defaultVariants: {
    width: 'default',
    height: 'default',
    maxWidth: 'default',
  },
});
