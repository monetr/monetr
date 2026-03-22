import type { ReactNode } from 'react';

import Flex from '@monetr/interface/components/Flex';
import { layoutVariants } from '@monetr/interface/components/Layout';
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
      <div className='p-0 md:p-8 w-full'>
        <MStepper activeIndex={step} steps={['Intro', 'Account', 'Balances', 'Income']} />
      </div>
      <Flex align='center' className={layoutVariants({ maxWidth: 'small' })} justify='center' orientation='column'>
        <MLogo className={layoutVariants({ size: 'logo' })} />
        {props.children}
      </Flex>
      {viewContext.metadata.showLogoutFooter && <LogoutFooter />}
      {!viewContext.metadata.showLogoutFooter && <div />}
    </div>
  );
}
