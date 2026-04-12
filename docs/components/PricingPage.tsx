import GradientHeading from '@monetr/docs/components/GradientHeading/GradientHeading';
import PricingCards from '@monetr/docs/components/Pricing/PricingCards';
import mergeClasses from '@monetr/docs/util/mergeClasses';

import styles from './PricingPage.module.scss';

export default function PricingPage(): JSX.Element {
  return (
    <div className={mergeClasses(styles.root, 'm-view-height')}>
      <div aria-hidden='true' className={styles.bgGlow}>
        <div className={styles.bgGlowRing}>
          <div className={styles.bgGlowOrb} />
        </div>
      </div>

      <div className={mergeClasses(styles.headerBlock, 'm-view-width')}>
        <div className={styles.titleBlock}>
          <GradientHeading
            blurClassName={styles.titleBlur}
            foregroundClassName={styles.titleForeground}
            wrapperClassName={styles.titleWrapper}
          >
            Pricing
          </GradientHeading>
        </div>
      </div>

      <div className={styles.pricingCardsWrap}>
        <PricingCards />
      </div>

      <div className={styles.faqBlock}>
        <h2 className={styles.faqHeading}>Frequently Asked Questions</h2>
        <ul className={styles.faqList}>
          <li>
            <h4 className={styles.faqQuestion}>Can I cancel anytime?</h4>
            <p>
              Yes. You can cancel at any time. You will have access to your account until the end of the current billing
              period.
            </p>
          </li>
          <li>
            <h4 className={styles.faqQuestion}>Is a payment method required to try monetr?</h4>
            <p>
              monetr does not require a payment method to try it out, however we do limit you to a single Plaid
              connection during the trial to try to prevent spam.
            </p>
          </li>
          <li>
            <h4 className={styles.faqQuestion}>What if I sign up and find monetr doesn't meet my needs?</h4>
            <p>
              If you have not subscribed yet then you don't need to do anything! Your account will become inactive
              automatically at the end of your trial. If you have already subscribed then you can cancel your
              subscription at any time!
            </p>
          </li>
        </ul>
      </div>
    </div>
  );
}
