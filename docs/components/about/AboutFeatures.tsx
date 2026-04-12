import GradientHeading from '@monetr/docs/components/GradientHeading/GradientHeading';

import styles from './AboutFeatures.module.scss';

export function AboutFeatures() {
  return (
    <div className={styles.root}>
      <GradientHeading
        blurClassName={styles.titleBlur}
        foregroundClassName={styles.titleForeground}
        wrapperClassName={styles.titleWrapper}
      >
        Coming Soon
      </GradientHeading>
    </div>
  );
}
