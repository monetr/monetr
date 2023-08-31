/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import { AccountBalance } from '@mui/icons-material';
import { Tooltip } from '@mui/material';

import { useBankAccounts, useSelectedBankAccount } from 'hooks/bankAccounts';
import { useInstitution } from 'hooks/institutions';
import MonetrLink from 'models/Link';
import mergeTailwind from 'util/mergeTailwind';

interface BankSidebarItemProps {
  link: MonetrLink;
}

export default function BankSidebarItem({ link }: BankSidebarItemProps): JSX.Element {
  const { result: institution } = useInstitution(link.plaidInstitutionId);
  const selectBankAccount = useSelectedBankAccount();
  const { data: bankAccounts } = useBankAccounts();
  const active = selectBankAccount.data?.linkId === link.linkId;

  const destinationBankAccountId = bankAccounts
    ?.find(bankAccount => bankAccount.linkId === link.linkId);

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

  return (
    <Tooltip title={ link.getName() } arrow placement='right' classes={ {
      tooltip: 'text-base font-medium',
    } }>
      <div
        className='w-full h-12 flex items-center justify-center relative group'
        data-testid={ `bank-sidebar-item-${link.linkId}` }
      >
        <div className={ classes } />
        <Link
          className='cursor-pointer absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle drop-shadow-md flex justify-center items-center'
          to={ `/bank/${destinationBankAccountId?.bankAccountId}/transactions` }
        >
          <InstitutionLogo />
        </Link>
      </div>
    </Tooltip>
  );
}
