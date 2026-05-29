import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Divider.module.scss';

export interface DividerProps {
  className?: string;
}

export default function Divider(props: DividerProps): JSX.Element {
  return <hr className={mergeClasses(styles.divider, props.className)} />;
}
