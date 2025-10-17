

import ManualLinkSetupAccountName from '@monetr/interface/components/setup/manual/ManualLinkSetupAccountName';
import ManualLinkSetupBalances from '@monetr/interface/components/setup/manual/ManualLinkSetupBalances';
import ManualLinkSetupIncome from '@monetr/interface/components/setup/manual/ManualLinkSetupIncome';
import ManualLinkSetupIntroName from '@monetr/interface/components/setup/manual/ManualLinkSetupIntroName';
import ManualLinkSetupLayout from '@monetr/interface/components/setup/manual/ManualLinkSetupLayout';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { ViewManager } from '@monetr/interface/components/ViewManager';

export interface ManualLinkSetupMetadata {
  showLogoutFooter?: boolean;
}

interface ManualLinkSetupProps extends ManualLinkSetupMetadata {}

export default function ManualLinkSetup(props: ManualLinkSetupProps): JSX.Element {
  const initialView: ManualLinkSetupSteps = ManualLinkSetupSteps.IntroName;

  return (
    <ViewManager<ManualLinkSetupSteps, ManualLinkSetupMetadata>
      initialView={initialView}
      initialMetadata={{
        showLogoutFooter: props.showLogoutFooter,
      }}
      viewComponents={{
        [ManualLinkSetupSteps.IntroName]: ManualLinkSetupIntroName,
        [ManualLinkSetupSteps.AccountName]: ManualLinkSetupAccountName,
        [ManualLinkSetupSteps.Balances]: ManualLinkSetupBalances,
        [ManualLinkSetupSteps.Income]: ManualLinkSetupIncome,
      }}
      layout={ManualLinkSetupLayout}
    />
  );
}
