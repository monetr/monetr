import { ArrowRight, CircleCheck } from 'lucide-react';

import styles from './PricingCards.module.scss';

import { Link } from '@rspress/core/theme-original';

export default function PricingCards(): JSX.Element {
  return (
    <div className={styles.root}>
      <div className={styles.cardFree}>
        <div className={styles.cardHeader}>
          <h3 className={styles.cardTitleFree}>You Host It</h3>
        </div>
        <div className={styles.cardPrice}>
          <span className={styles.cardPriceFreeValue}>Free Forever</span>
        </div>
        <ul className={styles.cardFeatureListFree}>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>Keep Your Data On Your Devices</span>
          </li>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>Community Support</span>
          </li>
        </ul>
        <Link className={styles.cardCtaFree} href='/documentation/install'>
          Get Started Now
        </Link>
      </div>

      <div className={styles.cardPaid}>
        <div className={styles.cardHeaderPaid}>
          <div>
            <h3 className={styles.cardTitlePaid}>We Host It</h3>
          </div>
        </div>
        <div className={styles.cardPricePaid}>
          <span className={styles.cardPriceValuePaid}>$4/month</span>
          <sup className={styles.cardPriceNotePaid}>*USD, tax included in price</sup>
        </div>
        <ul className={styles.cardFeatureListPaid}>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>30 Day Free Trial</span>
          </li>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>Access Anywhere Via Web</span>
          </li>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>Automatic Updates Via Plaid</span>
          </li>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>Import Via OFX</span>
          </li>
          <li className={styles.cardFeatureItem}>
            <CircleCheck className={styles.cardFeatureIcon} />
            <span>Email Support</span>
          </li>
        </ul>
        <Link className={styles.cardCtaPaid} href='https://my.monetr.app/register'>
          Try Free for 30 Days
          <ArrowRight />
        </Link>
      </div>
    </div>
  );
}
