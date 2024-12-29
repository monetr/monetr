import Feature from '@monetr/docs/components/Feature';
import BalancesScreenshot from '@monetr/docs/pages/assets/balances_hero.png';

export default function FreeToUse(): JSX.Element {
  return (
    <Feature
      title='See exactly what you have'
      className='col-span-2'
      image={ BalancesScreenshot }
    />
  );
}
