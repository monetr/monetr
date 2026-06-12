import React, { useState } from 'react';
import { CircleCheck, Pencil } from 'lucide-react';
import { Redirect } from 'wouter';

import Logo from '@monetr/interface/assets/Logo';
import { Button } from '@monetr/interface/components/Button';
import Flex, { flexVariants } from '@monetr/interface/components/Flex';
import LunchFlowLogo from '@monetr/interface/components/Logo/LunchFlowLogo';
import PlaidLogo from '@monetr/interface/components/Logo/PlaidLogo';
import LogoutFooter from '@monetr/interface/components/setup/LogoutFooter';
import SetupBillingButton from '@monetr/interface/components/setup/SetupBillingButton';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './setup.module.scss';

export interface SetupPageProps {
  // TODO Remove this prop and instead just use "does the user have any links". If they do then we can assume this is
  // true and if they don't we can treat this as false.
  alreadyOnboarded?: boolean;
  manualEnabled?: boolean;
}

type Step = 'greeting' | 'plaid' | 'lunchflow' | 'manual' | 'loading';

export default function SetupPage(props: SetupPageProps): React.JSX.Element {
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
      return <Redirect to={plaidPath} />;
    case 'lunchflow':
      return <Redirect to={lunchFlowPath} />;
    case 'manual':
      return <Redirect to={manualPath} />;
    default:
      return <h1>Something went wrong...</h1>;
  }
}

interface GreetingProps {
  alreadyOnboarded?: boolean;
  manualEnabled: boolean;
  onContinue: (_: Step) => unknown;
}

function Greeting(props: GreetingProps): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const [active, setActive] = useState<'plaid' | 'lunchflow' | 'manual' | null>(null);

  function Banner(): React.JSX.Element {
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
      className={mergeClasses(
        styles.greeting,
        flexVariants({
          flex: 'grow',
          align: 'center',
          gap: 'lg',
          justify: 'center',
          orientation: 'column',
        }),
      )}
    >
      <Logo className={styles.logo} />
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
          description='Connect to EU/UK/Global institutions via Lunch Flow.'
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

function OnboardingTile(props: OnboardingTileProps): React.JSX.Element {
  const disabledState = props.comingSoon || props.disabled;
  const wrapperClasses = mergeClasses(styles.tile, {
    [styles.tileActive]: !disabledState && props.active,
    [styles.tileInactive]: !disabledState && !props.active,
    [styles.tileDisabled]: disabledState,
  });

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
      {props.active && <CircleCheck className={styles.checkIcon} />}
      {React.createElement(props.icon, {
        className: styles.tileIcon,
      })}
      <div className={styles.tileBody}>
        <Typography size='lg' weight='medium'>
          {props.name}
        </Typography>
        <Typography color='subtle'>{props.description}</Typography>
        {!props.comingSoon && <Typography className={styles.spacerDesktop}>&nbsp;</Typography>}
        {props.comingSoon && (
          <Typography className={styles.tileFootnote} weight='medium'>
            Coming Soon
          </Typography>
        )}
        {props.disabled && (
          <Typography className={styles.tileFootnote} weight='medium'>
            Unavailable
          </Typography>
        )}
      </div>
    </button>
  );
}
