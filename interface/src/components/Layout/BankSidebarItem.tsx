/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';

import PlaidInstitutionLogo from '@monetr/interface/components/Plaid/InstitutionLogo';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import MonetrLink from '@monetr/interface/models/Link';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import sortAccounts from '@monetr/interface/util/sortAccounts';

interface BankSidebarItemProps {
  link: MonetrLink;
}

export default function BankSidebarItem({ link }: BankSidebarItemProps): JSX.Element {
  const selectBankAccount = useSelectedBankAccount();
  const { data: bankAccounts } = useBankAccounts();
  const active = selectBankAccount.data?.linkId === link.linkId;

  const destinationBankAccounts = sortAccounts(bankAccounts
    ?.filter(bankAccount => bankAccount.linkId === link.linkId));

  const destinationBankAccount = destinationBankAccounts.length > 0 ? destinationBankAccounts[0] : null;

  const LinkWarningIndicator = () => {
    const isWarning = link.getIsError() || link.getIsPendingExpiration();
    if (!isWarning) return null;

    return (
      <span className='absolute flex h-3 w-3 right-0 bottom-0'>
        <span className='animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-yellow-400' />
        <span className='relative inline-flex rounded-full h-3 w-3 bg-yellow-500' />
      </span>
    );
  };

  const LinkRevokedIndicator = () => {
    const isBad = link.getIsPlaid() && link.getIsRevoked();
    if (!isBad) return null;

    return (
      <span className='absolute flex h-3 w-3 right-0 bottom-0'>
        <span className='animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-red-400' />
        <span className='relative inline-flex rounded-full h-3 w-3 bg-red-500' />
      </span>
    );
  };

  const classes = mergeTailwind(
    'absolute',
    'dark:bg-dark-monetr-border',
    'right-0',
    'rounded-l-xl',
    'transition-transform',
    'w-1.5',
    {
      'h-8': active,
      'scale-y-100': active,
    },
    {
      'h-4': !active,
      'group-hover:scale-y-100': !active,
      'group-hover:scale-x-100': !active,
      'scale-x-0': !active,
      'scale-y-50': !active,
    },
  );

  let tooltip: string = link.getName();
  if (link.getIsPlaid()) {
    if (link.getIsError()) {
      tooltip = `${tooltip} (Error)`;
    } else if (link.getIsPendingExpiration()) {
      tooltip = `${tooltip} (Pending Expiration)`;
    } else if (link.getIsRevoked()) {
      tooltip = `${tooltip} (Disconnected)`;
    }
  }

  return (
    <Tooltip delayDuration={ 100 }>
      <TooltipTrigger
        className='w-full h-12 flex items-center justify-center relative group'
        data-testid={ `bank-sidebar-item-${link.linkId}` }
      >
        <div className={ classes } />
        <Link
          className='absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle drop-shadow-md flex justify-center items-center'
          to={ `/bank/${destinationBankAccount?.bankAccountId}/transactions` }
        >
          <PlaidInstitutionLogo link={ link } />
          <LinkWarningIndicator />
          <LinkRevokedIndicator />
        </Link>
      </TooltipTrigger>
      <TooltipContent side='right'>
        { tooltip }
      </TooltipContent>
    </Tooltip>
  );
}
