import Typography, { type TypographyProps } from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Badge.module.scss';

export interface BadgeProps extends Omit<TypographyProps, 'color'> {}

export default function Badge(props: BadgeProps): JSX.Element {
  return <Typography {...props} className={mergeTailwind(styles.badge, props.className)} />;
}
