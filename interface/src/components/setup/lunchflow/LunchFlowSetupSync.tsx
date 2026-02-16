import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import Typography from '@monetr/interface/components/Typography';

export default function LunchFlowSetupSync(): React.JSX.Element {
  return (
    <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
      <Typography align='center'>Loading...</Typography>
      <Typography align='center'>Getting your accounts setup, this can take a few seconds...</Typography>
    </LunchFlowSetupLayout>
  );
}
