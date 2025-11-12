import { useCallback } from 'react';

import { Button } from '@monetr/interface/components/Button';
import Flex from '@monetr/interface/components/Flex';
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
        <Button type='submit' variant='primary'>
          Next
        </Button>
      );
    case lastStep:
      return (
        <Flex gap='lg' justify='center'>
          <Button onClick={previousStep} variant='secondary'>
            Back
          </Button>
          <Button type='submit' variant='primary'>
            Finish
          </Button>
        </Flex>
      );
    default:
      return (
        <Flex gap='lg' justify='center'>
          <Button onClick={previousStep} variant='secondary'>
            Back
          </Button>
          <Button type='submit' variant='primary'>
            Next
          </Button>
        </Flex>
      );
  }
}
