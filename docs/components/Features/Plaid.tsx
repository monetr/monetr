import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function Plaid(): JSX.Element {
  return (
    <FeatureCard
      description='Using Plaid, monetr can receive secure automated updates from your bank. You never need to manually import your transactions or balances.'
      link='/documentation/use/plaid/'
      linkText='Learn About Plaid'
      title='Connect Your Bank With Plaid'
    />
  );
}
