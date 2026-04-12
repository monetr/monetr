import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function Expenses(): JSX.Element {
  return (
    <FeatureCard
      description={`monetr let's you budget for things that happen on all kinds of intervals, not confining you to a single monthly budget.`}
      link='/documentation/use/expense/'
      linkText='Learn More About Expenses'
      title='Track Recurring Expenses'
    />
  );
}
