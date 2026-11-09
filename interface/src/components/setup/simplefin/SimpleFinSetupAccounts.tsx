import { useCallback, useEffect, useMemo } from 'react';
import type { FormikErrors, FormikHelpers } from 'formik';
import { useLocation, useParams } from 'wouter';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormButton from '@monetr/interface/components/FormButton';
import SimpleFinLinkErrors from '@monetr/interface/components/link/SimpleFinLinkErrors';
import MForm from '@monetr/interface/components/MForm';
import SimpleFinSetupAccountItem from '@monetr/interface/components/setup/simplefin/SimpleFinSetupAccountItem';
import SimpleFinSetupLayout from '@monetr/interface/components/setup/simplefin/SimpleFinSetupLayout';
import { SimpleFinSetupSteps } from '@monetr/interface/components/setup/simplefin/SimpleFinSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import { useCreateLink } from '@monetr/interface/hooks/useCreateLink';
import { useSimpleFinBankAccounts } from '@monetr/interface/hooks/useSimpleFinBankAccounts';
import useSimpleFinBankAccountsRefresh from '@monetr/interface/hooks/useSimpleFinBankAccountsRefresh';
import { useSimpleFinLink } from '@monetr/interface/hooks/useSimpleFinLink';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { SimpleFinBankAccountStatus } from '@monetr/interface/models/SimpleFinBankAccount';
import { SimpleFinLinkStatus } from '@monetr/interface/models/SimpleFinLink';
import { useSnackbar } from '@monetr/notify';

export interface SimpleFinSetupAccountsForm {
  items: { [key: string]: boolean };
}

export default function SimpleFinSetupAccounts(): React.JSX.Element {
  const { simpleFinLinkId } = useParams<{ simpleFinLinkId: string }>();
  const [, navigate] = useLocation();
  const createLink = useCreateLink();
  const createBankAccount = useCreateBankAccount();
  const { enqueueSnackbar } = useSnackbar();
  // When the page loads, use the ID from the url params to trigger a refresh of bank accounts. This refresh will fail
  // if the ID is not valid which will prevent subsequent requests from happening. This refresh also populates the bank
  // account list in the API the first time it is called so it is necessary for this page.
  const {
    data: idToFetch,
    mutateAsync: refreshAccounts,
    isPending: isRefreshing,
    error: refreshError,
    isError: isErrorRefreshing,
    isSuccess: isRefreshComplete,
  } = useSimpleFinBankAccountsRefresh();
  // Once the hook above has completed, it will return an ID to fetch which we can then pass to the link hook and to the
  // bank accounts hook in order to proceed.
  const { data: simpleFinLink, isLoading: isLoadingLink, isError: isErrorLink } = useSimpleFinLink(idToFetch);
  const {
    data: simpleFinAccounts,
    isLoading: isLoadingAccounts,
    isError: isErrorAccounts,
  } = useSimpleFinBankAccounts(idToFetch);

  // Trigger the actual refresh as soon as we mount the page to make sure everything fetches!
  useEffect(() => void refreshAccounts(simpleFinLinkId), [simpleFinLinkId, refreshAccounts]);

  // If the simplefin link is in an active status then this page should not work!
  useEffect(() => {
    if (simpleFinLink && simpleFinLink?.status !== SimpleFinLinkStatus.Pending) {
      enqueueSnackbar('SimpleFIN link is not in a pending status and cannot be setup this way. Redirecting you...', {
        variant: 'warning',
        disableWindowBlurListener: true,
      });
      navigate('/');
    }
  }, [simpleFinLink, navigate, enqueueSnackbar]);

  const submit = useCallback(
    async (values: SimpleFinSetupAccountsForm, helpers: FormikHelpers<SimpleFinSetupAccountsForm>) => {
      // We can only submit once the link and accounts have loaded, the form is not rendered until then but bail just
      // in case.
      if (!simpleFinLink || !simpleFinAccounts) {
        return Promise.resolve();
      }

      helpers.setSubmitting(true);
      // Create the link from the simplefin link, this will move the simplefin link's status from pending to
      // active.
      return await createLink({
        institutionName: simpleFinLink.name,
        description: `Created via SimpleFIN`,
        simpleFinLinkId: simpleFinLink.simpleFinLinkId,
      })
        .then(result =>
          Promise.all(
            simpleFinAccounts
              .filter(account => values.items[account.simpleFinBankAccountId])
              .map(item =>
                createBankAccount({
                  linkId: result.linkId,
                  name: item.name,
                  simpleFinBankAccountId: item.simpleFinBankAccountId,
                  currency: item.currency,
                  currentBalance: item.currentBalance,
                  availableBalance: item.availableBalance,
                  accountType: BankAccountType.Depository,
                  accountSubType: BankAccountSubType.Checking,
                }),
              ),
          ),
        )
        .then(accounts => navigate(`/link/create/simplefin/${accounts[0]?.linkId}/sync`))
        .finally(() => helpers.setSubmitting(false));
    },
    [createBankAccount, createLink, simpleFinAccounts, simpleFinLink, navigate],
  );

  const validate = useCallback((values: SimpleFinSetupAccountsForm): FormikErrors<{ items: unknown }> => {
    const errors: FormikErrors<{ items: unknown }> = {};
    if (!Object.entries(values.items).some(([_, enabled]) => enabled)) {
      errors.items = 'Must enable at least one account.';
    }

    return errors;
  }, []);

  const initialValues: SimpleFinSetupAccountsForm = useMemo(
    () =>
      (simpleFinAccounts ?? []).reduce<SimpleFinSetupAccountsForm>(
        (acc, item) => {
          acc.items[item.simpleFinBankAccountId] = item.status === SimpleFinBankAccountStatus.Inactive;
          return acc;
        },
        { items: {} },
      ),
    [simpleFinAccounts],
  );

  if (isErrorLink || isErrorAccounts) {
    return (
      <SimpleFinSetupLayout step={SimpleFinSetupSteps.Accounts}>
        <Typography align='center'>Failed to retrieve SimpleFIN link details...</Typography>
      </SimpleFinSetupLayout>
    );
  }

  if (isErrorRefreshing) {
    return (
      <SimpleFinSetupLayout step={SimpleFinSetupSteps.Accounts}>
        <Typography align='center'>
          Failed to fetch accounts from SimpleFIN, please check your SimpleFIN account...
        </Typography>
        <Typography component='code'>{refreshError?.response?.data?.error}</Typography>
      </SimpleFinSetupLayout>
    );
  }

  // If we are loading ANY of our things, or if we have not started refreshing our data then we should show a loading
  // state. This should be the first few renders.
  if (
    isLoadingLink ||
    isLoadingAccounts ||
    isRefreshing ||
    !isRefreshComplete ||
    !simpleFinLink ||
    !simpleFinAccounts
  ) {
    return (
      <SimpleFinSetupLayout step={SimpleFinSetupSteps.Accounts}>
        <Typography align='center'>
          This can take several seconds while initial data is retrieved from SimpleFIN...
        </Typography>
        <Typography align='center'>Loading...</Typography>
      </SimpleFinSetupLayout>
    );
  }

  if (simpleFinAccounts.length === 0) {
    return (
      <SimpleFinSetupLayout step={SimpleFinSetupSteps.Accounts}>
        <Typography align='center'>No SimpleFIN accounts found.</Typography>
        <Typography align='center'>You might need to add connections to your account in SimpleFIN Bridge.</Typography>
        <SimpleFinLinkErrors errors={simpleFinLink.lastErrors} />
      </SimpleFinSetupLayout>
    );
  }

  if (simpleFinLink.status !== SimpleFinLinkStatus.Pending) {
    return (
      <SimpleFinSetupLayout step={SimpleFinSetupSteps.Accounts}>
        <Typography align='center'>SimpleFIN link is not in a pending status.</Typography>
        <Typography align='center'>You cannot proceed to setup accounts on this link...</Typography>
      </SimpleFinSetupLayout>
    );
  }

  return (
    <SimpleFinSetupLayout step={SimpleFinSetupSteps.Accounts}>
      <MForm
        className={flexVariants({ orientation: 'column', shrink: 'default', gap: 'lg' })}
        initialValues={initialValues}
        onSubmit={submit}
        validate={validate}
      >
        <Typography align='center'>
          monetr has found the following accounts in SimpleFIN, please select the accounts you want monetr to import.
        </Typography>
        <SimpleFinLinkErrors errors={simpleFinLink.lastErrors} />
        <ul>
          {simpleFinAccounts?.map(item => (
            <SimpleFinSetupAccountItem data={item} key={item.simpleFinBankAccountId} />
          ))}
        </ul>
        <div className={flexVariants({ justify: 'center' })}>
          <FormButton type='submit' variant='primary'>
            Continue
          </FormButton>
        </div>
      </MForm>
    </SimpleFinSetupLayout>
  );
}
