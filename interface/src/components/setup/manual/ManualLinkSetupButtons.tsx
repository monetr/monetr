import { useCallback } from 'react';

import { Button } from '@monetr/interface/components/Button';
import type { ManualLinkSetupForm } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';

export default function ManualLinkSetupButtons(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown, ManualLinkSetupForm>();
  const steps = Object.values(ManualLinkSetupSteps);
  const step = steps.indexOf(viewContext.currentView);
  const lastStep = steps.length - 1;

  const previousStep = useCallback(() => {
    viewContext.goToView(steps[step - 1]);
  }, [steps, step, viewContext]);

  switch (step) {
    case 0:
      return (
        <Button variant='primary' type='submit'>
          Next
        </Button>
      );
    case lastStep:
      return (
        <div className='flex gap-4'>
          <Button variant='secondary' onClick={previousStep}>
            Back
          </Button>
          <Button variant='primary' type='submit'>
            Finish
          </Button>
        </div>
      );
    default:
      return (
        <div className='flex gap-4'>
          <Button variant='secondary' onClick={previousStep}>
            Back
          </Button>
          <Button variant='primary' type='submit'>
            Next
          </Button>
        </div>
      );
  }
}
