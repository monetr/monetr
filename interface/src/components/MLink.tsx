import type React from 'react';
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

const colorMap: Record<NonNullable<MLinkProps['color']>, string> = {
  primary: styles.colorPrimary,
  secondary: styles.colorSecondary,
};

const sizeMap: Record<TextSize, string> = {
  xs: styles.sizeXs,
  sm: styles.sizeSm,
  md: styles.sizeMd,
  lg: styles.sizeLg,
  xl: styles.sizeXl,
  '2xl': styles.size2Xl,
  '3xl': styles.size3Xl,
  '4xl': styles.size4Xl,
  '5xl': styles.size5Xl,
};

export default function MLink(props: MLinkProps): JSX.Element {
  props = {
    ...MLinkPropsDefaults,
    ...props,
  };

  const classNames = mergeClasses(styles.root, colorMap[props.color], sizeMap[props.size], props.className);

  return (
    <Link {...props} className={classNames}>
      {props.children}
    </Link>
  );
}
