import { ArrowRight } from 'lucide-react';

import styles from './SignIn.module.scss';

import { Link } from '@rspress/core/theme-original';

export default function SignIn(): JSX.Element {
  return (
    <Link className={styles.link} href='https://my.monetr.app/'>
      <span className={styles.inner}>
        Sign In
        <ArrowRight className={styles.arrow} />
      </span>
    </Link>
  );
}
