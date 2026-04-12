import FeatureCard from '@monetr/docs/components/Features/FeatureCard';

export default function FileUpload(): JSX.Element {
  return (
    <FeatureCard
      description='Upload OFX files directly from your bank account to make it easy to get data into monetr.'
      link='/documentation/use/transactions/uploads/'
      linkText='Learn About Transaction Uploads'
      title='Import Transactions Manually'
    />
  );
}
