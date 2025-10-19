import { useEffect, useState } from 'react';
import * as Sentry from '@sentry/react';
import { useQueryClient } from '@tanstack/react-query';
import {
  type PlaidLinkError,
  type PlaidLinkOnExitMetadata,
  type PlaidLinkOnSuccessMetadata,
  type PlaidLinkOptionsWithLinkToken,
  usePlaidLink,
} from 'react-plaid-link';
import { useNavigate } from 'react-router-dom';

import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSpan from '@monetr/interface/components/MSpan';
import MSpinner from '@monetr/interface/components/MSpinner';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import type { ReactElement } from '@monetr/interface/components/types';
import request from '@monetr/interface/util/request';

interface PlaidProps {
  alreadyOnboarded?: boolean;
}

export default function PlaidSetup(props: PlaidProps): JSX.Element {
  interface State {
    token: string | null;
    loading: boolean;
    settingUp: boolean;
    error: string | null;
    exited: boolean;
  }

  const [{ token, loading, error, exited, settingUp }, setState] = useState<Partial<State>>({
    error: null,
    exited: false,
    loading: false,
    settingUp: false,
    token: null,
  });

  const queryClient = useQueryClient();
  const navigate = useNavigate();

  async function longPollSetup(recur: number, linkId: number): Promise<void> {
    setState({
      token,
      loading,
      error,
      exited,
      settingUp: true,
    });

    if (recur > 6) {
      return Promise.resolve();
    }

    return void request()
      .get(`/plaid/link/setup/wait/${linkId}`)
      .catch(error => {
        if (error.response.status === 408) {
          return longPollSetup(recur + 1, linkId);
        }

        throw error;
      });
  }

  function onPlaidExit(error: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) {
    if (error) {
      console.warn('Plaid link exited with error', {
        'plaid.request_id': metadata.request_id,
        'plaid.link_session_id': metadata.link_session_id,
        data: error,
      });
      Sentry.captureEvent({
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
      setState({
        token,
        loading,
        exited,
        error: 'Plaid link exited with an error.',
      });
    } else {
      setState({
        token,
        loading,
        exited: true,
      });
    }
  }

  async function onPlaidSuccess(public_token: string, metadata: PlaidLinkOnSuccessMetadata) {
    return request()
      .post('/plaid/link/token/callback', {
        publicToken: public_token,
        institutionId: metadata.institution.institution_id,
        institutionName: metadata.institution.name,
        accountIds: metadata.accounts.map((account: { id: string }) => account.id),
      })
      .then(async result => {
        const linkId: number = result.data.linkId;
        await longPollSetup(0, linkId);

        setTimeout(() => {
          queryClient.invalidateQueries({ queryKey: ['/links'] });
          queryClient.invalidateQueries({ queryKey: ['/bank_accounts'] });
          navigate('/');
        }, 8000);
      })
      .catch(error => {
        setState({
          token,
          error,
          loading: false,
          settingUp: false,
        });
      });
  }

  const config: PlaidLinkOptionsWithLinkToken = {
    token: token,
    onSuccess: onPlaidSuccess,
    onExit: onPlaidExit,
    onLoad: null,
    onEvent: null,
  };

  const { error: plaidError, open: plaidOpen } = usePlaidLink(config);
  useEffect(() => {
    request()
      .get('/plaid/link/token/new?use_cache=true')
      .then(result =>
        setTimeout(() => {
          setState({
            loading: false,
            token: result.data.linkToken,
            error: null,
          });
        }, 1000),
      )
      .catch(error =>
        setTimeout(() => {
          const message = error?.response?.data?.error || 'Could not connect to Plaid, an unknown error occurred.';
          setState({
            loading: false,
            token: null,
            error: message,
          });
        }, 3000),
      );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (token && plaidOpen) {
      plaidOpen();
      setState({
        token,
        loading: true,
      });
    }
  }, [token, plaidOpen]);

  useEffect(() => {
    if (plaidError) {
      Sentry.captureException(plaidError);
    }
  }, [plaidError]);

  let inner: ReactElement = (
    <div className='flex flex-col justify-center items-center'>
      <MSpan className='text-2xl font-medium'>Getting Plaid Ready!</MSpan>
      <MSpan className='text-lg' color='subtle'>
        One moment while we prepare your connection with Plaid.
      </MSpan>
    </div>
  );

  if (settingUp) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>Successfully connected with Plaid!</MSpan>
        <MSpan className='text-lg' color='subtle'>
          Hold on a moment while we pull the initial data from Plaid into monetr.
        </MSpan>
      </div>
    );
  }

  if (loading) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpinner />
      </div>
    );
  }

  if (error) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>Something isn't quite right</MSpan>
        <MSpan className='text-lg' color='subtle'>
          We were not able to get Plaid ready for you at this time, please try again shortly.
        </MSpan>
        <MSpan className='text-lg' color='subtle'>
          If the problem continues, please contact support@monetr.app
        </MSpan>
        <MSpan className='text-md' color='muted'>
          Error Message: {error}
        </MSpan>
      </div>
    );
  }

  if (exited) {
    const backUrl = props.alreadyOnboarded ? '/link/create' : '/setup';
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan size='2xl' weight='medium'>
          Something isn't quite right
        </MSpan>
        <MSpan size='lg' color='subtle'>
          Plaid exited, did you want to set it up later?
        </MSpan>
        <MSpan size='md' color='subtle'>
          Or <MLink to={backUrl}>go back</MLink> and pick another option?
        </MSpan>
      </div>
    );
  }

  function Footer(): JSX.Element {
    if (props.alreadyOnboarded) {
      return null;
    }

    return <LogoutFooter />;
  }

  return (
    <div className='w-full h-full flex justify-center items-center gap-8 flex-col overflow-hidden text-center p-2'>
      <MLogo className='w-24 h-24' />
      {inner}
      <Footer />
    </div>
  );
}
