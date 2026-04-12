import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function FreeToUse(): JSX.Element {
  return (
    <FeatureCard
      description='monetr keeps track of how much you have put aside for your budgets so it can tell you exactly how much you have left over to use or spend.'
      link='/documentation/use/free_to_use/'
      linkText='Learn About Free-To-Use'
      title={`See What's Leftover`}
    />
  );
}
