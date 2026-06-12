import type React from 'react';
import { cva } from 'class-variance-authority';
import { Link } from 'wouter';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './MLink.module.scss';
import type { TextSize } from './types';

export interface MLinkProps extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> {
  to: string;
  children: React.ReactNode;
  color?: 'primary' | 'secondary';
  size?: TextSize;
}

const MLinkPropsDefaults: Omit<MLinkProps, 'children' | 'to'> = {
  size: 'md',
  color: 'primary',
  tabIndex: 0,
};

const linkVariants = cva([styles.root], {
  variants: {
    color: {
      primary: styles.colorPrimary,
      secondary: styles.colorSecondary,
    },
    size: {
      xs: styles.sizeXs,
      sm: styles.sizeSm,
      md: styles.sizeMd,
      lg: styles.sizeLg,
      xl: styles.sizeXl,
      '2xl': styles.size2Xl,
      '3xl': styles.size3Xl,
      '4xl': styles.size4Xl,
      '5xl': styles.size5Xl,
    },
  },
});

export default function MLink(props: MLinkProps): React.JSX.Element {
  props = {
    ...MLinkPropsDefaults,
    ...props,
  };

  const classNames = mergeClasses(linkVariants({ color: props.color, size: props.size }), props.className);

  return (
    <Link {...props} className={classNames}>
      {props.children}
    </Link>
  );
}
