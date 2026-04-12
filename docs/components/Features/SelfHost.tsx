import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function SelfHost(): JSX.Element {
  return (
    <FeatureCard
      description='Host monetr yourself on your own hardware, for free. Keeping your data private.'
      link='/documentation/install/'
      linkText='Installation Guide'
      title='Self-Host'
    />
  );
}
