import type React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Card.module.scss';

export interface CardProps extends React.HTMLProps<HTMLDivElement> {}

export default function Card({ className, ...props }: CardProps): JSX.Element {
  return <div className={mergeTailwind(styles.card, className)} {...props} />;
}
