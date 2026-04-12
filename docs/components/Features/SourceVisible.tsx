import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function SourceVisible(): JSX.Element {
  return (
    <FeatureCard
      description={`All of monetr's source code is publically available, you can see exactly how we handle your data and even contribute functionality!`}
      link='https://github.com/monetr/monetr'
      linkExternal
      linkText='See The Source Code'
      title='Source Visible'
    />
  );
}
