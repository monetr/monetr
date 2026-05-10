import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';

import styles from './ArrowLink.module.scss';

export interface ArrowRedirectProps {
  to: string;
}

export default function ArrowLink(props: ArrowRedirectProps): JSX.Element {
  return (
    <Link className={styles.arrowLink} tabIndex={-1} to={props.to}>
      <ChevronRight />
    </Link>
  );
}
