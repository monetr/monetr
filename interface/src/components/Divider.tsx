import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Divider.module.scss';

export interface DividerProps {
  className?: string;
}

export default function Divider(props: DividerProps): JSX.Element {
  return <hr className={mergeTailwind(styles.divider, props.className)} />;
}
