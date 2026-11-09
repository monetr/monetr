import type { ReactNode } from 'react';
import { Plus } from 'lucide-react';

import { flexVariants } from '@monetr/interface/components/Flex';
import SimpleFinLogo from '@monetr/interface/components/Logo/SimpleFinLogo';
import MLogo from '@monetr/interface/components/MLogo';
import MStepper from '@monetr/interface/components/MStepper';
import { SimpleFinSetupSteps } from '@monetr/interface/components/setup/simplefin/SimpleFinSetupSteps';

import styles from './SimpleFinSetupLayout.module.scss';

interface SimpleFinSetupLayoutProps {
  step?: SimpleFinSetupSteps;
  children: ReactNode;
}

export default function SimpleFinSetupLayout(props: SimpleFinSetupLayoutProps): React.JSX.Element {
  const steps = Object.values(SimpleFinSetupSteps);
  const step = steps.indexOf(props.step ?? SimpleFinSetupSteps.Intro);
  return (
    <div className={styles.simpleFinSetupLayoutRoot}>
      <div className={styles.stepper}>
        <MStepper activeIndex={step} steps={['Intro', 'Accounts', 'Sync']} />
      </div>
      <div className={styles.content}>
        <div className={flexVariants({ gap: 'sm', width: 'auto' })}>
          <MLogo className={styles.logo} />
          <Plus className={styles.icon} />
          <SimpleFinLogo className={styles.logo} />
        </div>
        {props.children}
      </div>
      <div />
    </div>
  );
}
