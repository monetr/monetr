import FormTextField from '@monetr/interface/components/FormTextField';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

export default function BetaCodeInput(): JSX.Element {
  const { data: config } = useAppConfiguration();
  if (!config?.requireBetaCode) {
    return null;
  }

  return (
    <FormTextField
      className='w-full md:w-1/2 lg:w-1/3 xl:w-1/4'
      label='Beta Code'
      name='betaCode'
      required
      type='text'
      uppercasetext
    />
  );
}
