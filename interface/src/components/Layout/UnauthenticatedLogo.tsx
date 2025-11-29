import MLogo from '@monetr/interface/components/MLogo';

import styles from './UnauthenticatedLogo.module.scss';

export default function UnauthenticatedLogo(): React.JSX.Element {
  return (
    <div className={styles.logo}>
      <MLogo />
    </div>
  );
}
