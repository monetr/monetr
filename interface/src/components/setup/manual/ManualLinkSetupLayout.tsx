import React, { ReactNode } from 'react';

import MLogo from '@monetr/interface/components/MLogo';
import MStepper from '@monetr/interface/components/MStepper';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import { ManualLinkSetupMetadata } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';

interface ManualLinkSetupLayoutProps {
  children: ReactNode | undefined;
}

export default function ManualLinkSetupLayout(props: ManualLinkSetupLayoutProps): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, ManualLinkSetupMetadata>();
  const steps = Object.values(ManualLinkSetupSteps);
  const step = steps.indexOf(viewContext.currentView);
  return (
    <div className='w-full h-full flex justify-between items-center gap-8 flex-col p-4 md:p-2 overflow-auto' >
      <div className='p-0 md:p-8 w-full'>
        <MStepper steps={ ['Intro', 'Account', 'Balances', 'Income'] } activeIndex={ step } />
      </div>
      <div className='flex flex-col md:justify-center items-center max-w-sm h-full' >
        <MLogo className='w-24 h-24' />
        { props.children }
      </div>
      { viewContext.metadata.showLogoutFooter && 
        <LogoutFooter />
      }
    </div>
  );
}
