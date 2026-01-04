import type { ReactNode } from 'react';

import MLogo from '@monetr/interface/components/MLogo';
import MStepper from '@monetr/interface/components/MStepper';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import type {
  ManualLinkSetupForm,
  ManualLinkSetupMetadata,
} from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';

import styles from './ManualLinkSetupLayout.module.scss';

interface ManualLinkSetupLayoutProps {
  children: ReactNode | undefined;
}

export default function ManualLinkSetupLayout(props: ManualLinkSetupLayoutProps): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, ManualLinkSetupMetadata, ManualLinkSetupForm>();
  const steps = Object.values(ManualLinkSetupSteps);
  const step = steps.indexOf(viewContext.currentView);
  return (
    <div className={styles.manualLinkSetupLayoutRoot}>
      <div className='p-0 md:p-2 w-full'>
        <MStepper activeIndex={step} steps={['Intro', 'Account', 'Balances', 'Income']} />
      </div>
      <div className='flex flex-col md:justify-center items-center max-w-sm h-full'>
        <MLogo className='w-24 h-24' />
        {props.children}
      </div>
      {viewContext.metadata.showLogoutFooter && <LogoutFooter />}
    </div>
  );
}
