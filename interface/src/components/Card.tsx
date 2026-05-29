import type React from 'react';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Card.module.scss';

export interface CardProps extends React.HTMLProps<HTMLDivElement> {}

export default function Card({ className, ...props }: CardProps): JSX.Element {
  return <div className={mergeClasses(styles.card, className)} {...props} />;
}
