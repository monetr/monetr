import { useEffect } from 'react';
import { useLocation, useParams } from 'wouter';

import { Button } from '@monetr/interface/components/Button';
import { flexVariants } from '@monetr/interface/components/Flex';
import SimpleFinSetupLayout from '@monetr/interface/components/setup/simplefin/SimpleFinSetupLayout';
import { SimpleFinSetupSteps } from '@monetr/interface/components/setup/simplefin/SimpleFinSetupSteps';
import SimpleFinSetupSyncItem from '@monetr/interface/components/setup/simplefin/SimpleFinSetupSyncItem';
import Typography from '@monetr/interface/components/Typography';
import { useBankAccountsForLink } from '@monetr/interface/hooks/useBankAccountsForLink';
import useSimpleFinLinkSyncProgress from '@monetr/interface/hooks/useSimpleFinLinkSyncProgress';
import request from '@monetr/interface/util/request';

export default function SimpleFinSetupSync(): React.JSX.Element {
  const [, navigate] = useLocation();
  const { linkId } = useParams<{ linkId: string }>();
  const { data: bankAccounts } = useBankAccountsForLink(linkId);
  // This has to stay above the effect below, effects run in order and we want the socket opening before the sync starts
  const { data: progress } = useSimpleFinLinkSyncProgress(linkId);

  // As soon as this page loads immediately trigger the simplefin sync
  useEffect(
    () =>
      void request({
        method: 'POST',
        url: '/api/simplefin/link/sync',
        data: {
          linkId,
        },
      }),
    [linkId],
  );

  return (
    <SimpleFinSetupLayout step={SimpleFinSetupSteps.Sync}>
      <Typography align='center'>Getting your accounts setup, this can take a few seconds...</Typography>
      <ul>
        {bankAccounts?.map(item => (
          <SimpleFinSetupSyncItem bankAccount={item} key={item.bankAccountId} status={progress?.status} />
        ))}
      </ul>
      <div className={flexVariants({ justify: 'center' })}>
        <Button
          onClick={() => navigate(`/bank/${bankAccounts?.at(0)?.bankAccountId}/transactions`)}
          type='submit'
          variant='primary'
        >
          Continue
        </Button>
      </div>
    </SimpleFinSetupLayout>
  );
}
