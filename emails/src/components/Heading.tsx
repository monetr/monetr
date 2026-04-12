import type React from 'react';

import styles from './Heading.module.scss';

type HeadingAs = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

export type HeadingProps = React.ComponentPropsWithoutRef<'h1'> & {
  as?: HeadingAs;
};

export default function Heading({ as: Tag = 'h1', className, ...props }: HeadingProps) {
  return <Tag className={className || styles.heading} {...props} />;
}
