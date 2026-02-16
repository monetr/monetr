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
      <div className={flexVariants({ justify: 'center' })}>
        {/* TODO, Instead redirect to one of the accounts we just setup! */}
        <Button onClick={() => navigate('/')} type='submit' variant='primary'>
          Continue
        </Button>
      </div>
    </LunchFlowSetupLayout>
  );
}
