import { useMemo } from 'react';
import { useParams } from 'react-router-dom';

import LunchFlowSetupAccounts from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupAccounts';
import LunchFlowSetupIntro from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupIntro';
import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import LunchFlowSetupSync from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSync';
import { ViewManager } from '@monetr/interface/components/ViewManager';

interface LunchFlowSetupProps {
  className?: string;
}

export interface LunchFlowSetupMetadata extends LunchFlowSetupProps {
  lunchFlowLinkId?: string;
}

export type LunchFlowSetupForm = unknown;

export default function LunchFlowSetup(props: LunchFlowSetupProps): React.JSX.Element {
  const { lunchFlowLinkId } = useParams();
  // If we already have a link ID then we don't need to create one and we skip right to the accounts setup page.
  const initialView = useMemo(
    () => (lunchFlowLinkId ? LunchFlowSetupSteps.Accounts : LunchFlowSetupSteps.Intro),
    [lunchFlowLinkId],
  );

  return (
    <ViewManager<LunchFlowSetupSteps, LunchFlowSetupMetadata, LunchFlowSetupForm>
      initialMetadata={{
        ...props,
        lunchFlowLinkId,
      }}
      initialView={initialView}
      layout={LunchFlowSetupLayout}
      viewComponents={{
        [LunchFlowSetupSteps.Intro]: LunchFlowSetupIntro,
        [LunchFlowSetupSteps.Accounts]: LunchFlowSetupAccounts,
        [LunchFlowSetupSteps.Sync]: LunchFlowSetupSync,
      }}
    />
  );
}
