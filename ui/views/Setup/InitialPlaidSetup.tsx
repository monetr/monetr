import React, { Fragment, useState } from 'react';
import { PlaidLinkError, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types/index';
import { useQueryClient } from 'react-query';
import { Button, Typography } from '@mui/material';
import { Severity } from '@sentry/react';
import * as Sentry from '@sentry/react';

import { Logo } from 'assets';
import PlaidButton from 'components/Plaid/PlaidButton';
import PlaidIcon from 'components/Plaid/PlaidIcon';
import { useAppConfiguration } from 'hooks/useAppConfiguration';
import useLogout from 'hooks/useLogout';
import request from 'util/request';

interface State {
  linkId: number | null;
  loading: boolean;
  longPollAttempts: number;
}

const InitialSetupBilling = (): JSX.Element => {
  const {
    billingEnabled,
  } = useAppConfiguration();

  async function manageSubscription() {
    return request().get('/billing/portal')
      .then(result => {
        window.location.assign(result.data.url);
        return Promise.resolve();
      })
      .catch(error => {
        alert(error);
      });
  }

  if (!billingEnabled) {
    return null;
  }

  return (
    <Fragment>
      <div className="w-full opacity-50 pt-2.5 pb-2.5">
        <div className="relative w-full border-t border-gray-400 top-5" />
        <div className="relative flex justify-center inline w-full">
          <span className="relative bg-white p-1.5">
              or
          </span>
        </div>
      </div>
      <div className="w-full pt-2.5 pb-2.5">
        <Button
          onClick={ manageSubscription }
          color="secondary"
          className="w-full"
        >
          Manage your subscription
        </Button>
      </div>
    </Fragment>
  );
};

export default function InitialPlaidSetup(): JSX.Element {
  const [state, setState] = useState<Partial<State>>({
    loading: false,
  });
  const queryClient = useQueryClient();

  async function longPollSetup(linkId: number): Promise<void> {
    setState({
      loading: true,
      longPollAttempts: state.longPollAttempts + 1,
    });

    const { longPollAttempts } = state;
    if (longPollAttempts > 6) {
      return Promise.resolve();
    }

    return void request().get(`/plaid/link/setup/wait/${ linkId }`)
      .catch(error => {
        if (error.response.status === 408) {
          return longPollSetup(linkId);
        }

        throw error;
      });
  }

  function onPlaidExit(error: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) {
    if (error) {
      console.warn('Plaid link exited with error', {
        'plaid.request_id': metadata.request_id,
        'plaid.link_session_id': metadata.link_session_id,
        'data': error,
      });
      Sentry.captureEvent({
        message: 'Plaid link exited with error',
        level: Severity.Error,
        tags: {
          'plaid.request_id': metadata.request_id,
          'plaid.link_session_id': metadata.link_session_id,
        },
        breadcrumbs: [
          {
            type: 'info',
            level: Severity.Info,
            message: 'Error from Plaid link',
            data: error,
          },
        ],
      });
    }
  }

  async function onPlaidSuccess(public_token: string, metadata: PlaidLinkOnSuccessMetadata) {
    setState({
      loading: true,
    });

    return request().post('/plaid/link/token/callback', {
      publicToken: public_token,
      institutionId: metadata.institution.institution_id,
      institutionName: metadata.institution.name,
      accountIds: metadata.accounts.map((account: { id: string }) => account.id),
    })
      .then(result => {
        const linkId: number = result.data.linkId;
        setState({
          linkId: linkId,
          loading: true,
        });

        return longPollSetup(linkId)
          .then(() => Promise.all([
            queryClient.invalidateQueries('/links'),
            queryClient.invalidateQueries('/bank_accounts'),
          ]));
      })
      .catch(error => {
        setState({
          loading: false,
        });

        throw error;
      });
  }

  const logout = useLogout();
  console.log('testing things');

  return (
    <div className="flex justify-center w-full h-full max-h-full pb-5">
      <div className="flex flex-col w-full h-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
        <div className="flex items-center justify-center flex-grow">
          <div>
            <div className="flex justify-center w-full mb-5">
              <img src={ Logo } className="w-1/3" />
            </div>
            <div className="w-full pt-2.5 pb-2.5">
              <Typography
                className="w-full text-center"
              >
                monetr uses Plaid to retrieve bank account data automatically.
              </Typography>
            </div>
            <div className="w-full pt-2.5 pb-2.5">
              <PlaidButton
                className="w-full"
                color="primary"
                disabled={ state.loading }
                plaidOnSuccess={ onPlaidSuccess }
                plaidOnExit={ onPlaidExit }
                variant="outlined"
              >
                Get Started with
                <PlaidIcon className="flex-none w-16 ml-2" />
              </PlaidButton>
            </div>
            <InitialSetupBilling />
          </div>
        </div>
        <div className="flex-initial w-full pt-2.5 pb-2.5">
          <Button
            onClick={ logout }
            className="w-full opacity-50"
          >
            Logout
          </Button>
        </div>
      </div>
    </div>
  );
};
