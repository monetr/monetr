import { ChevronRight } from 'lucide-react';
import { Link } from 'wouter';

import styles from './ArrowLink.module.scss';

export interface ArrowRedirectProps {
  to: string;
}

export default function ArrowLink(props: ArrowRedirectProps): React.JSX.Element {
  return (
    <Link className={styles.arrowLink} tabIndex={-1} to={props.to}>
      <ChevronRight />
    </Link>
  );
}
