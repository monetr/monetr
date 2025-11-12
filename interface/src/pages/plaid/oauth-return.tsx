import { useEffect, useState } from 'react';
import { captureEvent } from '@sentry/react';
import { useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import MSpan from '@monetr/interface/components/MSpan';
import MSpinner from '@monetr/interface/components/MSpinner';
import { OAuthRedirectPlaidLink } from '@monetr/interface/components/Plaid/OAuthRedirectPlaidLink';
import request from '@monetr/interface/util/request';

import type { PlaidLinkError, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types';

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
  linkId: number | null;
  longPollAttempts: number;
}

export default function OauthReturn(): JSX.Element {
  const queryClient = useQueryClient();
  const [state, setState] = useState<Partial<State>>({
    loading: true,
  });

  useEffect(() => {
    request()
      .get('/plaid/link/token/new?use_cache=true')
      .then(result =>
        setState({
          loading: false,
          linkToken: result.data.linkToken,
        }),
      )
      .catch(error =>
        setState({
          loading: false,
          error: error,
        }),
      );
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

    return request()
      .get(`/plaid/link/setup/wait/${linkId}`)
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
      captureEvent({
        message: 'Plaid link exited with error',
        level: 'error',
        tags: {
          'plaid.request_id': metadata.request_id,
          'plaid.link_session_id': metadata.link_session_id,
        },
        breadcrumbs: [
          {
            type: 'info',
            level: 'info',
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

    return void request()
      .post('/plaid/link/token/callback', {
        publicToken: public_token,
        institutionId: metadata.institution.institution_id,
        institutionName: metadata.institution.name,
        accountIds: metadata.accounts.map((account: { id: string }) => account.id),
      })
      .then(result => {
        setState({
          linkId: result.data.linkId,
        });

        return longPollSetup().then(() =>
          Promise.all([
            queryClient.invalidateQueries({ queryKey: ['/links'] }),
            queryClient.invalidateQueries({ queryKey: ['/bank_accounts'] }),
          ]),
        );
      });
  }

  function renderContents(): JSX.Element {
    if (state.loading || !state.linkToken) {
      return (
        <div>
          <MSpan size='xl'>One moment...</MSpan>
          <div className='flex flex-col justify-center items-center'>
            <MSpinner />
          </div>
        </div>
      );
    }

    return (
      <div>
        <OAuthRedirectPlaidLink
          linkToken={state.linkToken}
          plaidOnExit={plaidLinkExit}
          plaidOnSuccess={plaidLinkSuccess}
        />
      </div>
    );
  }

  return (
    <div className='w-full h-full flex justify-center items-center p-10'>
      <div>{renderContents()}</div>
    </div>
  );
}
