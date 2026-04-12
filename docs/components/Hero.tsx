import { ArrowRight } from 'lucide-react';

import Expenses from '@monetr/docs/components/Features/Expenses';
import FileUpload from '@monetr/docs/components/Features/FileUpload';
import Forecasting from '@monetr/docs/components/Features/Forecasting';
import FreeToUse from '@monetr/docs/components/Features/FreeToUse';
import MobileFriendly from '@monetr/docs/components/Features/MobileFriendly';
import Plaid from '@monetr/docs/components/Features/Plaid';
import SelfHost from '@monetr/docs/components/Features/SelfHost';
import SourceVisible from '@monetr/docs/components/Features/SourceVisible';
import GradientHeading from '@monetr/docs/components/GradientHeading/GradientHeading';
import ScreenshotCarousel from '@monetr/docs/components/ScreenshotCarousel';
import mergeClasses from '@monetr/docs/util/mergeClasses';

import styles from './Hero.module.scss';

import { Link } from '@rspress/core/theme-original';

export default function Hero(): JSX.Element {
  return (
    <div className={mergeClasses(styles.root, 'm-view-height')}>
      <div aria-hidden='true' className={styles.bgGlow}>
        <div className={styles.bgGlowRing}>
          <div className={styles.bgGlowOrb} />
        </div>
      </div>

      <div className={mergeClasses(styles.content, 'm-view-width')}>
        <div className={styles.titleBlock}>
          <GradientHeading
            blurClassName={styles.titleBlur}
            foregroundClassName={styles.titleForeground}
            wrapperClassName={styles.titleWrapper}
          >
            Welcome to monetr
          </GradientHeading>

          <h1 className={styles.subtitle}>Take control of your finances, paycheck by paycheck</h1>

          <h2 className={styles.tagline}>Put aside what you need, spend what you want.</h2>
        </div>

        <div className={styles.ctaRow}>
          <Link className={styles.ctaPrimary} href='https://my.monetr.app/register'>
            Try Free for 30 Days
            <ArrowRight />
          </Link>
          <Link className={styles.ctaSecondary} href='/documentation/use/getting_started'>
            Learn More
          </Link>
        </div>

        <ScreenshotCarousel />

        <h1 className={styles.featuresTitle}>Features</h1>

        <div className={styles.featuresGrid}>
          <FreeToUse />
          <Expenses />
          <FileUpload />
          <Plaid />
          <Forecasting />
          <MobileFriendly />
          <SelfHost />
          <SourceVisible />
        </div>
      </div>
    </div>
  );
}
