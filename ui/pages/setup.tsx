/* eslint-disable max-len */
import React, { useEffect, useState } from 'react';
import { PlaidLinkError, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata, PlaidLinkOptionsWithLinkToken, usePlaidLink } from 'react-plaid-link';
import { useQueryClient } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { CheckCircle, EditOutlined, LinkOutlined } from '@mui/icons-material';
import * as Sentry from '@sentry/react';

import { MBaseButton } from 'components/MButton';
import MLink from 'components/MLink';
import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import MSpinner from 'components/MSpinner';
import { ReactElement } from 'components/types';
import mergeTailwind from 'util/mergeTailwind';
import request from 'util/request';

export interface SetupPageProps {
  manualEnabled?: boolean;
}

type Step = 'greeting'|'plaid'|'manual'|'loading';

export default function SetupPage(props: SetupPageProps): JSX.Element {
  const [step, setStep] = useState<Step>('greeting');

  switch (step) {
    case 'greeting':
      return <Greeting onContinue={ setStep } manualEnabled={ props.manualEnabled } />;
    case 'plaid':
      return <Plaid />;
    case 'manual':
      // Not implemented yet.
      return null;
    case 'loading':

    default:
      return <h1>Something went wrong...</h1>;
  }
}

interface GreetingProps {
  manualEnabled: boolean;
  onContinue: (_: Step) => unknown;
}

function Greeting(props: GreetingProps): JSX.Element {
  const [active, setActive] = useState<'plaid'|'manual'|null>(null);

  // TODO Make this more compatible with portrait mode mobile.
  return (
    <div className='w-full h-full flex justify-center items-center gap-8 flex-col overflow-hidden text-center'>
      <MLogo className='w-24 h-24' />
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>
          Welcome to monetr!
        </MSpan>
        <MSpan className='text-lg' variant='light'>
          Before we get started, please select how you would like to continue.
        </MSpan>
      </div>
      <div className='flex gap-4'>
        <OnboardingTile
          icon={ <LinkOutlined /> }
          name='Connected'
          description='Connect your bank account using Plaid.'
          active={ active === 'plaid' }
          onClick={ () => setActive('plaid') }
        />
        <OnboardingTile
          icon={ <EditOutlined /> }
          name='Manual'
          description='Enter transaction and balance data manually.'
          active={ active === 'manual' }
          onClick={ () => setActive('manual') }
          comingSoon={ !props.manualEnabled }
        />
      </div>
      <MBaseButton
        color={ !active ? 'secondary' : 'primary' }
        disabled={ !active }
        onClick={ () => props.onContinue(active) }
      >
        Continue
      </MBaseButton>
      <div className='flex justify-center gap-1'>
        <MSpan variant="light" className='text-sm'>Not ready to continue?</MSpan>
        <MLink to="/logout" size="sm">Logout for now</MLink>
      </div>
    </div>
  );
}

interface OnboardingTileProps {
  onClick: () => void;
  active: boolean;
  icon: React.ReactElement;
  name: ReactElement;
  description: ReactElement;
  comingSoon?: boolean;
}

function OnboardingTile(props: OnboardingTileProps): JSX.Element {
  const nonDisabled = mergeTailwind(
    {
      'dark:border-dark-monetr-brand': props.active,
      'dark:hover:border-dark-monetr-brand-subtle': props.active,
      'border-monetr-brand': props.active,
      'hover:border-monetr-brand-subtle': props.active,
    },
    {
      'dark:border-dark-monetr-border': !props.active,
      'dark:hover:border-dark-monetr-border-string': !props.active,
      'border-monetr-border': !props.active,
      'hover:border-monetr-border-string': !props.active,
    },
    'cursor-pointer',
    'border'
  );
  const disabled = mergeTailwind(
    'cursor-not-allowed',
    'dark:ring-dark-monetr-border-subtle',
    'ring-monetr-border-subtle',
    'ring-1',
    'ring-inset',
    'dark:text-dark-monetr-content-muted',
    'text-monetr-content-muted',
    'opacity-50',
  );

  const wrapperClasses = mergeTailwind(
    { [nonDisabled]: !props.comingSoon },
    { [disabled]: props.comingSoon },
    'flex',
    'flex-col',
    'gap-4',
    'group',
    'h-72',
    'items-center',
    'p-4',
    'relative',
    'rounded-lg',
    'w-56',
  );

  function handleClick() {
    if (props.comingSoon) return;

    props.onClick();
  }

  return (
    <div className={ wrapperClasses } onClick={ handleClick }>
      { props.active && <CheckCircle className='absolute dark:text-dark-monetr-brand-subtle top-2 right-2' /> }
      { React.cloneElement(props.icon, { className: 'w-16 h-16' }) }
      <div className='flex flex-col gap-2 items-center h-full mt-4'>
        <MSpan className='text-lg font-medium'>
          { props.name }
        </MSpan>
        <MSpan variant='light'>
          { props.description }
        </MSpan>
        { props.comingSoon &&
          <MSpan className='mt-5 font-medium'>
            Coming Soon
          </MSpan>
        }
      </div>
    </div>
  );
}

function Plaid(): JSX.Element {
  interface State {
    token: string | null;
    loading: boolean;
    settingUp: boolean;
    error: string | null;
  }

  const [{ token, loading, error, settingUp }, setState] = useState<Partial<State>>({
    token: null,
    loading: false,
    settingUp: false,
    error: null,
  });

  const queryClient = useQueryClient();
  const navigate = useNavigate();

  async function longPollSetup(recur: number, linkId: number): Promise<void> {
    setState({
      token,
      loading,
      error,
      settingUp: true,
    });

    if (recur > 6) {
      return Promise.resolve();
    }

    return void request().get(`/plaid/link/setup/wait/${ linkId }`)
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
        'data': error,
      });
      Sentry.captureEvent({
        message: 'Plaid link exited with error',
        level: Sentry.Severity.Error,
        tags: {
          'plaid.request_id': metadata.request_id,
          'plaid.link_session_id': metadata.link_session_id,
        },
        breadcrumbs: [
          {
            type: 'info',
            level: Sentry.Severity.Info,
            message: 'Error from Plaid link',
            data: error,
          },
        ],
      });
      setState({
        token,
        loading,
        error: 'Plaid link exited with an error.',
      });
    }
  }

  async function onPlaidSuccess(public_token: string, metadata: PlaidLinkOnSuccessMetadata) {
    return request().post('/plaid/link/token/callback', {
      publicToken: public_token,
      institutionId: metadata.institution.institution_id,
      institutionName: metadata.institution.name,
      accountIds: metadata.accounts.map((account: { id: string }) => account.id),
    })
      .then(async result => {
        const linkId: number = result.data.linkId;
        await longPollSetup(0, linkId);

        setTimeout(() => {
          queryClient.invalidateQueries('/links');
          queryClient.invalidateQueries('/bank_accounts');
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
    request().get('/plaid/link/token/new?use_cache=true')
      .then(result => setTimeout(() => {
        setState({
          loading: false,
          token: result.data.linkToken,
          error: null,
        });
      }, 1000))
      .catch(error => setTimeout(() => {
        const message = error?.response?.data?.error || 'Could not connect to Plaid, an unknown error occurred.';
        setState({
          loading: false,
          token: null,
          error: message,
        });
      }, 3000));
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
      <MSpan className='text-2xl font-medium'>
        Getting Plaid Ready!
      </MSpan>
      <MSpan className='text-lg' variant='light'>
        One moment while we prepare your connection with Plaid.
      </MSpan>
    </div>
  );

  if (settingUp) {
    inner = (
      <div className='flex flex-col justify-center items-center'>
        <MSpan className='text-2xl font-medium'>
          Successfully connected with Plaid!
        </MSpan>
        <MSpan className='text-lg' variant='light'>
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
        <MSpan className='text-2xl font-medium'>
          Something isn't quite right
        </MSpan>
        <MSpan className='text-lg' variant='light'>
          We were not able to get Plaid ready for you at this time, please try again shortly.
        </MSpan>
        <MSpan className='text-lg' variant='light'>
          If the problem continues, please contact support@monetr.app
        </MSpan>
      </div>
    );
  }

  return (
    <div className='w-full h-full flex justify-center items-center gap-8 flex-col overflow-hidden text-center p-2'>
      <MLogo className='w-24 h-24' />
      { inner }
      <div className='flex justify-center gap-1'>
        <MSpan variant="light" className='text-sm'>Not ready to continue?</MSpan>
        <MLink to="/logout" size="sm">Logout for now</MLink>
      </div>
    </div>
  );
}
