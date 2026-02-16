import { useParams } from 'react-router-dom';

import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useBankAccountsForLink } from '@monetr/interface/hooks/useBankAccountsForLink';
import { useEffect } from 'react';
import request from '@monetr/interface/util/request';
import LunchFlowSetupSyncItem from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSyncItem';

export default function LunchFlowSetupSync(): React.JSX.Element {
  const { linkId } = useParams();
  const { data: bankAccounts, isLoading } = useBankAccountsForLink(linkId);

  useEffect(() => {
    request().post(`/lunch_flow/link/sync`, {
      linkId,
    });
  }, [linkId]);

  return (
    <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
      <Typography align='center'>Loading...</Typography>
      <Typography align='center'>Getting your accounts setup, this can take a few seconds...</Typography>
      <ul>
        {bankAccounts?.map(item => (
          <LunchFlowSetupSyncItem bankAccount={item} key={item.bankAccountId} />
        ))}
      </ul>
    </LunchFlowSetupLayout>
  );
}
