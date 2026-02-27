import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { Button } from '@monetr/interface/components/Button';
import { flexVariants } from '@monetr/interface/components/Flex';
import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import LunchFlowSetupSyncItem from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSyncItem';
import Typography from '@monetr/interface/components/Typography';
import { useBankAccountsForLink } from '@monetr/interface/hooks/useBankAccountsForLink';
import request from '@monetr/interface/util/request';

export default function LunchFlowSetupSync(): React.JSX.Element {
  const navigate = useNavigate();
  const { linkId } = useParams();
  const { data: bankAccounts } = useBankAccountsForLink(linkId);

  // As soon as this page loads immediately trigger the lunch flow sync
  useEffect(
    () =>
      void request().post(`/lunch_flow/link/sync`, {
        linkId,
      }),
    [linkId],
  );

  return (
    <LunchFlowSetupLayout step={LunchFlowSetupSteps.Sync}>
      <Typography align='center'>Getting your accounts setup, this can take a few seconds...</Typography>
      <ul>
        {bankAccounts?.map(item => (
          <LunchFlowSetupSyncItem bankAccount={item} key={item.bankAccountId} />
        ))}
      </ul>
      <div className={flexVariants({ justify: 'center' })}>
        <Button
          onClick={() => navigate(`/bank/${bankAccounts?.at(0).bankAccountId}/transactions`)}
          type='submit'
          variant='primary'
        >
          Continue
        </Button>
      </div>
    </LunchFlowSetupLayout>
  );
}
