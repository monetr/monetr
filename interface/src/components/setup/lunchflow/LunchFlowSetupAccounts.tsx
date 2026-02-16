import { useEffect, useMemo } from 'react';
import { useParams } from 'react-router-dom';

import { Button } from '@monetr/interface/components/Button';
import { flexVariants } from '@monetr/interface/components/Flex';
import MForm from '@monetr/interface/components/MForm';
import LunchFlowSetupAccountItem from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupAccountItem';
import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useLunchFlowBankAccounts } from '@monetr/interface/hooks/useLunchFlowBankAccounts';
import useLunchFlowBankAccountsRefresh from '@monetr/interface/hooks/useLunchFlowBankAccountsRefresh';
import { useLunchFlowLink } from '@monetr/interface/hooks/useLunchFlowLink';

export interface LunchFlowSetupAccountsForm {
  items: { [key: string]: boolean };
}

export default function LunchFlowSetupAccounts(): React.JSX.Element {
  const { lunchFlowLinkId } = useParams();
  // When the page loads, use the ID from the url params to trigger a refresh of bank accounts. This refresh will fail
  // if the ID is not valid which will prevent subsequent requests from happening. This refresh also populates the bank
  // account list in the API the first time it is called so it is necessary for this page.
  const {
    data: idToFetch,
    mutateAsync: refreshAccounts,
    isPending: isRefreshing,
    isError: isErrorRefreshing,
    isSuccess: isRefreshComplete,
  } = useLunchFlowBankAccountsRefresh();
  // Once the hook above has completed, it will return an ID to fetch which we can then pass to the link hook and to the
  // bank accounts hook in order to proceed.
  const { isLoading: isLoadingLink, isError: isErrorLink } = useLunchFlowLink(idToFetch);
  const {
    data: lunchFlowAccounts,
    isLoading: isLoadingAccounts,
    isError: isErrorAccounts,
  } = useLunchFlowBankAccounts(idToFetch);

  // Trigger the actual refresh as soon as we mount the page to make sure everything fetches!
  useEffect(() => void refreshAccounts(lunchFlowLinkId), [lunchFlowLinkId, refreshAccounts]);

  const initialValues: LunchFlowSetupAccountsForm = useMemo(
    () =>
      (lunchFlowAccounts ?? []).reduce(
        (acc, item) => {
          acc.items[item.lunchFlowBankAccountId] = true;
          return acc;
        },
        { items: {} },
      ),
    [lunchFlowAccounts],
  );

  if (isErrorLink || isErrorAccounts) {
    return (
      <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
        <Typography align='center'>Failed to retrieve Lunch Flow link details...</Typography>
      </LunchFlowSetupLayout>
    );
  }

  if (isErrorRefreshing) {
    return (
      <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
        <Typography align='center'>
          Failed to fetch accounts from Lunch Flow, please check your API credentials...
        </Typography>
      </LunchFlowSetupLayout>
    );
  }

  // If we are loading ANY of our things, or if we have not started refreshing our data then we should show a loading
  // state. This should be the first few renders.
  if (isLoadingLink || isLoadingAccounts || isRefreshing || !isRefreshComplete) {
    return (
      <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
        <Typography align='center'>Loading...</Typography>
        <Typography align='center'>
          This can take several seconds while initial data is retrieved from Lunch Flow...
        </Typography>
      </LunchFlowSetupLayout>
    );
  }

  // TODO Add an empty state here too!
  return (
    <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
      <MForm className={flexVariants({ orientation: 'column' })} initialValues={initialValues} onSubmit={() => {}}>
        <Typography align='center'>
          monetr has found the following accounts in Lunch Flow, please select the accounts you want monetr to import.
        </Typography>
        <ul>
          {lunchFlowAccounts?.map(item => (
            <LunchFlowSetupAccountItem data={item} key={item.lunchFlowBankAccountId} />
          ))}
        </ul>
        <div className={flexVariants({ justify: 'center' })}>
          <Button variant='secondary'>Cancel</Button>
          <Button type='submit' variant='primary'>
            Continue
          </Button>
        </div>
      </MForm>
    </LunchFlowSetupLayout>
  );
}
