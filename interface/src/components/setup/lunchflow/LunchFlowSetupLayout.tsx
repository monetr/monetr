import type { ReactNode } from 'react';

import MLogo from '@monetr/interface/components/MLogo';
import MStepper from '@monetr/interface/components/MStepper';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';

import styles from './LunchFlowSetupLayout.module.scss';

interface LunchFlowSetupLayoutProps {
  children: ReactNode;
}

export default function LunchFlowSetupLayout(props: LunchFlowSetupLayoutProps): React.JSX.Element {
  const steps = Object.values(LunchFlowSetupSteps);
  const step = steps.indexOf(LunchFlowSetupSteps.Intro);
  return (
    <div className={styles.lunchFlowSetupLayoutRoot}>
      <div className='p-0 md:p-8 w-full'>
        <MStepper activeIndex={step} steps={['Intro', 'Accounts', 'Sync']} />
      </div>
      <div className='flex flex-col md:justify-center items-center max-w-sm'>
        <MLogo className='w-24 h-24' />
        {props.children}
      </div>
      <div />
    </div>
  );
}
