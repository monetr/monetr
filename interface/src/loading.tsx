import { flexVariants } from '@monetr/interface/components/Flex';
import { heights, widths } from '@monetr/interface/components/Layout';
import MLogo from '@monetr/interface/components/MLogo';
import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './loading.module.scss';

export default function Loading(): JSX.Element {
  return (
    <div
      className={mergeTailwind(
        flexVariants({ justify: 'center', align: 'center', orientation: 'column', gap: 'xl' }),
        heights.screen,
        widths.screen,
      )}
    >
      <MLogo className={styles.logo} />
      <div className={mergeTailwind(styles.blobs, flexVariants({ align: 'center', justify: 'center' }))}>
        <div className={styles.dot} />
        <div className={styles.dots} />
        <div className={styles.dots} />
        <div className={styles.dots} />
      </div>
      <Typography size='5xl'>Loading...</Typography>
      <svg className={styles.filter} version='1.1' xmlns='http://www.w3.org/2000/svg'>
        <title>Loading...</title>
        <defs>
          <filter id='magic'>
            <feGaussianBlur in='SourceGraphic' result='blur' stdDeviation='10' />
            <feColorMatrix in='blur' mode='matrix' result='goo' values='1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 18 -7' />
            <feBlend in='SourceGraphic' in2='goo' />
          </filter>
        </defs>
      </svg>
    </div>
  );
}
