import { Plug } from 'lucide-react';

import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import SetupPage from '@monetr/interface/pages/setup';

import { Fragment } from 'react/jsx-runtime';

export default function LinkCreatePage(): JSX.Element {
  const { data: config } = useAppConfiguration();

  return (
    <Fragment>
      <MTopNavigation icon={Plug} title='Add another connection' />
      <SetupPage alreadyOnboarded manualEnabled={config?.manualEnabled} />
    </Fragment>
  );
}
