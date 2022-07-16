import React, { useEffect, useState } from 'react';
import { PlaidLinkError, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types';
import { useQueryClient } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CircularProgress, Typography } from '@mui/material';
import * as Sentry from '@sentry/react';
import { Severity } from '@sentry/react';

import { OAuthRedirectPlaidLink } from 'components/Plaid/OAuthRedirectPlaidLink';
import request from 'util/request';

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
  linkId: number | null;
  longPollAttempts: number;
}

export default function OAuthRedirect(): JSX.Element {
  const queryClient = useQueryClient();
  const [state, setState] = useState<Partial<State>>({
    loading: true,
  });

  useEffect(() => {
    request()
      .get('/plaid/link/token/new?use_cache=true')
      .then(result => setState({
        loading: false,
        linkToken: result.data.linkToken,
      }))
      .catch(error => setState({
        loading: false,
        error: error,
      }));
  }, []);

  const navigate = useNavigate();

  async function longPollSetup(): Promise<void> {
    setState(prevState => ({
      longPollAttempts: prevState.longPollAttempts + 1,
    }));

    const { longPollAttempts, linkId } = state;
    if (longPollAttempts > 6) {
      return Promise.resolve();
    }

    return request().get(`/plaid/link/setup/wait/${ linkId }`)
      .then(() => Promise.resolve())
      .catch(error => {
        if (error.response.status === 408) {
          return longPollSetup();
        }

        throw error;
      });
  }

  function plaidLinkExit(error: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) {
    if (error) {
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

    return navigate('/');
  }

  async function plaidLinkSuccess(public_token: string, metadata: PlaidLinkOnSuccessMetadata): Promise<void> {
    setState({ loading: true });

    return void request().post('/plaid/link/token/callback', {
      publicToken: public_token,
      institutionId: metadata.institution.institution_id,
      institutionName: metadata.institution.name,
      accountIds: metadata.accounts.map((account: { id: string}) => account.id),
    })
      .then(result => {
        setState({
          linkId: result.data.linkId,
        });

        return longPollSetup()
          .then(() => Promise.all([
            queryClient.invalidateQueries('/links'),
            queryClient.invalidateQueries('/bank_accounts'),
          ]));
      });
  }

  function renderContents(): JSX.Element {
    if (state.loading || !state.linkToken) {
      return (
        <div>
          <Typography variant="h5">
            One moment...
          </Typography>
          <div className="flex justify-center items-center p-5 m-5">
            <CircularProgress />
          </div>
        </div>
      );
    }

    return (
      <div>
        <OAuthRedirectPlaidLink
          linkToken={ state.linkToken }
          plaidOnSuccess={ plaidLinkSuccess }
          plaidOnExit={ plaidLinkExit }
        />
      </div>
    );
  }

  return (
    <div className="w-full h-full flex justify-center items-center p-10">
      <div>
        <Card>
          <CardContent>
            { renderContents() }
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

