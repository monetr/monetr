import React, { useState } from 'react';
import { CircleCheck, Pencil } from 'lucide-react';
import { Navigate } from 'react-router-dom';

import Logo from '@monetr/interface/assets/Logo';
import { Button } from '@monetr/interface/components/Button';
import Flex, { flexVariants } from '@monetr/interface/components/Flex';
import { layoutVariants } from '@monetr/interface/components/Layout';
import LunchFlowLogo from '@monetr/interface/components/Logo/LunchFlowLogo';
import PlaidLogo from '@monetr/interface/components/Logo/PlaidLogo';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import SetupBillingButton from '@monetr/interface/components/setup/SetupBillingButton';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface SetupPageProps {
  // TODO Remove this prop and instead just use "does the user have any links". If they do then we can assume this is
  // true and if they don't we can treat this as false.
  alreadyOnboarded?: boolean;
  manualEnabled?: boolean;
}

type Step = 'greeting' | 'plaid' | 'lunchflow' | 'manual' | 'loading';

export default function SetupPage(props: SetupPageProps): JSX.Element {
  const [step, setStep] = useState<Step>('greeting');
  const plaidPath = props.alreadyOnboarded ? '/link/create/plaid' : '/setup/plaid';
  const manualPath = props.alreadyOnboarded ? '/link/create/manual' : '/setup/manual';
  const lunchFlowPath = props.alreadyOnboarded ? '/link/create/lunchflow' : '/setup/lunchflow';

  switch (step) {
    case 'greeting':
      return (
        <Greeting alreadyOnboarded={props.alreadyOnboarded} manualEnabled={props.manualEnabled} onContinue={setStep} />
      );
    case 'plaid':
      return <Navigate to={plaidPath} />;
    case 'lunchflow':
      return <Navigate to={lunchFlowPath} />;
    case 'manual':
      return <Navigate to={manualPath} />;
    default:
      return <h1>Something went wrong...</h1>;
  }
}

interface GreetingProps {
  alreadyOnboarded?: boolean;
  manualEnabled: boolean;
  onContinue: (_: Step) => unknown;
}

function Greeting(props: GreetingProps): JSX.Element {
  const { data: config } = useAppConfiguration();
  const [active, setActive] = useState<'plaid' | 'lunchflow' | 'manual' | null>(null);

  function Banner(): JSX.Element {
    if (!props.alreadyOnboarded) {
      return (
        <Flex align='center' gap='lg' justify='center' orientation='column'>
          <Typography align='center' size='2xl' weight='medium'>
            Welcome to monetr!
          </Typography>
          <Typography align='center' color='subtle' size='lg'>
            Before we get started, please select how you would like to continue.
          </Typography>
        </Flex>
      );
    }

    return (
      <Flex align='center' gap='lg' justify='center' orientation='column'>
        <Typography size='2xl' weight='medium'>
          Adding another bank?
        </Typography>
        <Typography align='center' color='subtle' size='lg'>
          Please select what type of bank you want to setup below.
        </Typography>
      </Flex>
    );
  }

  return (
    <div
      className={mergeTailwind(
        'p-4',
        { 'h-screen': !props.alreadyOnboarded },
        flexVariants({
          align: 'center',
          gap: 'lg',
          justify: 'center',
          orientation: 'column',
        }),
        layoutVariants({ width: 'full', height: 'full' }),
      )}
    >
      <Logo className='size-16 md:size-24' />
      <Banner />
      <Flex gap='lg' justify='center' orientation='stackMedium'>
        <OnboardingTile
          active={active === 'plaid'}
          description='Plaid makes connecting your monetr account to your bank easy.'
          disabled={!config?.plaidEnabled}
          icon={PlaidLogo}
          name='Plaid'
          onClick={() => setActive('plaid')}
        />
        <OnboardingTile
          active={active === 'lunchflow'}
          description='Connect to EU/UK institutions via Lunch Flow.'
          disabled={!config.lunchFlowEnabled}
          icon={LunchFlowLogo}
          name='Lunch Flow'
          onClick={() => setActive('lunchflow')}
        />
        <OnboardingTile
          active={active === 'manual'}
          description='Manage your transactions and budget manually with monetr.'
          disabled={!props.manualEnabled}
          icon={Pencil}
          name='Manual'
          onClick={() => setActive('manual')}
        />
      </Flex>
      <Button color={!active ? 'secondary' : 'primary'} disabled={!active} onClick={() => props.onContinue(active)}>
        Continue
      </Button>
      {!props.alreadyOnboarded && <SetupBillingButton />}
      {!props.alreadyOnboarded && <LogoutFooter />}
    </div>
  );
}

interface OnboardingTileProps {
  onClick: () => void;
  active: boolean;
  icon: React.FC<{ className?: string }>; // Anything that allows the class name to be customized
  name: React.ReactNode;
  description: React.ReactNode;
  comingSoon?: boolean;
  disabled?: boolean;
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
    'border',
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

  const disabledState = props.comingSoon || props.disabled;
  const wrapperClasses = mergeTailwind(
    { [nonDisabled]: !disabledState },
    { [disabled]: disabledState },
    'text-center',
    'flex',
    'flex-row',
    'md:flex-col',
    'gap-4',
    'group',
    'md:h-72',
    'md:w-56',
    'items-center',
    'p-2',
    'py-4',
    'md:p-4',
    'relative',
    'rounded-lg',
    'h-36',
  );

  function handleClick() {
    if (props.comingSoon) {
      return;
    }
    if (props.disabled) {
      return;
    }

    props.onClick();
  }

  if (props.disabled) {
    return null;
  }

  return (
    <button className={wrapperClasses} onClick={handleClick} type='button'>
      {props.active && <CircleCheck className='absolute dark:text-dark-monetr-brand-subtle top-2 right-2' />}
      {React.createElement(props.icon, {
        className: 'w-auto h-12 md:h-10 ml-4 md:ml-0 md:mt-6 dark:text-dark-monetr-content-emphasis',
      })}
      <div className='flex flex-col gap-2 items-center h-full md:mt-4 text-center w-full md:w-auto'>
        <Typography size='lg' weight='medium'>
          {props.name}
        </Typography>
        <Typography color='subtle'>{props.description}</Typography>
        {!props.comingSoon && <Typography className='md:block hidden'>&nbsp;</Typography>}
        {props.comingSoon && (
          <Typography className='md:mt-5' weight='medium'>
            Coming Soon
          </Typography>
        )}
        {props.disabled && (
          <Typography className='md:mt-5' weight='medium'>
            Unavailable
          </Typography>
        )}
      </div>
    </button>
  );
}
