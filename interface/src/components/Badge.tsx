import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Badge.module.scss';
import MSpan, { type MSpanProps } from './MSpan';

export interface BadgeProps extends Omit<MSpanProps, 'color'> {}

export default function Badge(props: BadgeProps): JSX.Element {
  return <MSpan {...props} className={mergeTailwind(styles.badge, props.className)} />;
}
