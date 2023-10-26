/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import { AccountBalance } from '@mui/icons-material';
import { Tooltip } from '@mui/material';

import { useBankAccounts, useSelectedBankAccount } from 'hooks/bankAccounts';
import { useInstitution } from 'hooks/institutions';
import MonetrLink from 'models/Link';
import mergeTailwind from 'util/mergeTailwind';
import sortAccounts from 'util/sortAccounts';

interface BankSidebarItemProps {
  link: MonetrLink;
}

export default function BankSidebarItem({ link }: BankSidebarItemProps): JSX.Element {
  const { result: institution } = useInstitution(link.plaidInstitutionId);
  const selectBankAccount = useSelectedBankAccount();
  const { data: bankAccounts } = useBankAccounts();
  const active = selectBankAccount.data?.linkId === link.linkId;

  const destinationBankAccounts = sortAccounts(bankAccounts
    ?.filter(bankAccount => bankAccount.linkId === link.linkId));

  const destinationBankAccount = destinationBankAccounts.length > 0 ? destinationBankAccounts[0] : null;

  const InstitutionLogo = () => {
    if (!institution?.logo) {
      return (
        <AccountBalance
          data-testid={ `bank-sidebar-item-${link.linkId}-logo-missing` }
          color='info'
        />
      );
    }

    return (
      <img
        data-testid={ `bank-sidebar-item-${link.linkId}-logo` }
        src={ `data:image/png;base64,${institution.logo}` }
      />
    );
  };

  const LinkWarningIndicator = () => {
    if (!link.getIsError()) return null;

    return (
      <span className="absolute flex h-3 w-3 right-0 bottom-0">
        <span className="animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-yellow-400" />
        <span className="relative inline-flex rounded-full h-3 w-3 bg-yellow-500" />
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
  if (link.getIsError()) {
    tooltip = `${tooltip} (Error)`;
  }

  return (
    <Tooltip title={ tooltip } arrow placement='right' classes={ {
      tooltip: 'text-base font-medium',
    } }>
      <div
        className='w-full h-12 flex items-center justify-center relative group'
        data-testid={ `bank-sidebar-item-${link.linkId}` }
      >
        <div className={ classes } />
        <Link
          className='absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle drop-shadow-md flex justify-center items-center'
          to={ `/bank/${destinationBankAccount?.bankAccountId}/transactions` }
        >
          <InstitutionLogo />
          <LinkWarningIndicator />
        </Link>
      </div>
    </Tooltip>
  );
}
