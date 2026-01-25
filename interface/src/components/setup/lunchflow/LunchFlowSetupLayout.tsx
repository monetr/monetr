import type { ReactNode } from 'react';
import { Plus } from 'lucide-react';

import LunchFlowLogo from '@monetr/interface/components/Logo/LunchFlowLogo';
import MLogo from '@monetr/interface/components/MLogo';
import MStepper from '@monetr/interface/components/MStepper';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';

import styles from './LunchFlowSetupLayout.module.scss';

interface LunchFlowSetupLayoutProps {
  step?: LunchFlowSetupSteps;
  children: ReactNode;
}

export default function LunchFlowSetupLayout(props: LunchFlowSetupLayoutProps): React.JSX.Element {
  const steps = Object.values(LunchFlowSetupSteps);
  const step = steps.indexOf(props.step ?? LunchFlowSetupSteps.Intro);
  return (
    <div className={styles.lunchFlowSetupLayoutRoot}>
      <div className='p-0 md:p-8 w-full'>
        <MStepper activeIndex={step} steps={['Intro', 'Accounts', 'Sync']} />
      </div>
      <div className='flex flex-col md:justify-center items-center max-w-sm gap-2'>
        <div className='flex gap-2'>
          <MLogo className='w-16 h-16' />
          <Plus className='h-16' />
          <LunchFlowLogo className='h-16' iconOnly />
        </div>
        {props.children}
      </div>
      <div />
    </div>
  );
}
