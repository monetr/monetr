import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function Forecasting(): JSX.Element {
  return (
    <FeatureCard
      description={`See a forecast of your finances based on your budget, so you can see how much you'll have and when you'll have it.`}
      link='/documentation/use/forecasting/'
      linkText='Learn About Forecasting'
      title='See Your Financial Future'
    />
  );
}
