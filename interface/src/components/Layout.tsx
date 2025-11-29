import { cva } from 'class-variance-authority';

import styles from './Layout.module.scss';

export const widths = {
  default: undefined,
  '1/4': styles.widthQuarter,
  '1/3': styles.widthThird,
  '1/2': styles.widthHalf,
  full: styles.widthFull,
  screen: styles.widthScreen,
};

export const heights = {
  default: undefined,
  screen: styles.heightScreen,
};

export const layoutVariants = cva([], {
  variants: {
    width: widths,
    height: heights,
  },
  defaultVariants: {
    width: 'default',
    height: 'default',
  },
});
