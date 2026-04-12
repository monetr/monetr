import type React from 'react';

import styles from './Link.module.scss';

export type LinkProps = React.ComponentPropsWithoutRef<'a'>;

export function Link({ target = '_blank', className, ...props }: LinkProps) {
  return <a className={className || styles.link} target={target} {...props} />;
}
