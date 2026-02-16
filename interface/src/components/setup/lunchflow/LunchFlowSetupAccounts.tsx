import { useCallback, useEffect, useMemo } from 'react';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useNavigate, useParams } from 'react-router-dom';

import { Button } from '@monetr/interface/components/Button';
import { flexVariants } from '@monetr/interface/components/Flex';
import FormButton from '@monetr/interface/components/FormButton';
import MForm from '@monetr/interface/components/MForm';
import LunchFlowSetupAccountItem from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupAccountItem';
import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import { useCreateLink } from '@monetr/interface/hooks/useCreateLink';
import { useLunchFlowBankAccounts } from '@monetr/interface/hooks/useLunchFlowBankAccounts';
import useLunchFlowBankAccountsRefresh from '@monetr/interface/hooks/useLunchFlowBankAccountsRefresh';
import { useLunchFlowLink } from '@monetr/interface/hooks/useLunchFlowLink';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { LunchFlowBankAccountStatus } from '@monetr/interface/models/LunchFlowBankAccount';

export interface LunchFlowSetupAccountsForm {
  items: { [key: string]: boolean };
}

export default function LunchFlowSetupAccounts(): React.JSX.Element {
  const { lunchFlowLinkId } = useParams();
  const navigate = useNavigate();
  const createLink = useCreateLink();
  const createBankAccount = useCreateBankAccount();
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
  // TODO If the lunch flow link is in an active status then this page should not work!
  const { data: lunchFlowLink, isLoading: isLoadingLink, isError: isErrorLink } = useLunchFlowLink(idToFetch);
  const {
    data: lunchFlowAccounts,
    isLoading: isLoadingAccounts,
    isError: isErrorAccounts,
  } = useLunchFlowBankAccounts(idToFetch);

  // Trigger the actual refresh as soon as we mount the page to make sure everything fetches!
  useEffect(() => void refreshAccounts(lunchFlowLinkId), [lunchFlowLinkId, refreshAccounts]);

  const submit = useCallback(
    async (values: LunchFlowSetupAccountsForm, helpers: FormikHelpers<LunchFlowSetupAccountsForm>) => {
      helpers.setSubmitting(true);
      // Create the link from the lunch flow link, this will move the lunch flow link's status from pending to
      // active.
      return createLink({
        institutionName: lunchFlowLink?.name,
        description: `Created via Lunch Flow`,
        lunchFlowLinkId: lunchFlowLink?.lunchFlowLinkId,
      })
        .then(result =>
          Promise.all(
            lunchFlowAccounts
              .filter(account => values.items[account.lunchFlowBankAccountId])
              .map(item =>
                createBankAccount({
                  linkId: result.linkId,
                  name: item.name,
                  lunchFlowBankAccountId: item.lunchFlowBankAccountId,
                  currency: item.currency,
                  currentBalance: item.currentBalance,
                  availableBalance: item.currentBalance,
                  accountType: BankAccountType.Depository,
                  accountSubType: BankAccountSubType.Checking,
                }),
              ),
          ),
        )
        .then(() =>
          navigate('sync', {
            relative: 'path',
          }),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [createBankAccount, createLink, lunchFlowAccounts, lunchFlowLink, navigate],
  );

  const validate = useCallback((values: LunchFlowSetupAccountsForm): FormikErrors<{ items: unknown }> => {
    const errors: FormikErrors<{ items: unknown }> = {};
    if (!Object.entries(values.items).some(([_, enabled]) => enabled)) {
      errors.items = 'Must enable at least one account.';
    }

    return errors;
  }, []);

  const initialValues: LunchFlowSetupAccountsForm = useMemo(
    () =>
      (lunchFlowAccounts ?? []).reduce(
        (acc, item) => {
          acc.items[item.lunchFlowBankAccountId] = item.status === LunchFlowBankAccountStatus.Inactive;
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

  if (lunchFlowAccounts.length === 0) {
    return (
      <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
        <Typography align='center'>No Lunch Flow accounts found.</Typography>
        <Typography align='center'>You might need to add connections to your API destination in Lunch Flow.</Typography>
      </LunchFlowSetupLayout>
    );
  }

  return (
    <LunchFlowSetupLayout step={LunchFlowSetupSteps.Accounts}>
      <MForm
        className={flexVariants({ orientation: 'column', shrink: 'default', gap: 'lg' })}
        initialValues={initialValues}
        onSubmit={submit}
        validate={validate}
      >
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
          <FormButton type='submit' variant='primary'>
            Continue
          </FormButton>
        </div>
      </MForm>
    </LunchFlowSetupLayout>
  );
}
