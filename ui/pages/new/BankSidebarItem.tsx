/* eslint-disable max-len */
import React from 'react';
import { AccountBalance } from '@mui/icons-material';
import shallow from 'zustand/shallow';

import { useBankAccountsSink, useSelectedBankAccount } from 'hooks/bankAccounts';
import { useInstitution } from 'hooks/institutions';
import useStore from 'hooks/store';
import Link from 'models/Link';
import mergeTailwind from 'util/mergeTailwind';

interface BankSidebarItemProps {
  link: Link;
}

export default function BankSidebarItem({ link }: BankSidebarItemProps): JSX.Element {
  const { result: institution } = useInstitution(link.plaidInstitutionId);
  const { bankAccount } = useSelectedBankAccount();
  const { result: bankAccounts } = useBankAccountsSink();
  const { setCurrentBankAccount } = useStore(state => ({
    setCurrentBankAccount: state.setCurrentBankAccount,
  }), shallow);
  const active = bankAccount?.linkId === link.linkId;

  function onClick() {
    const newSelectedBankAccount = Array.from(bankAccounts.values()).find(bankAccount => bankAccount.linkId === link.linkId);
    if (newSelectedBankAccount?.bankAccountId) {
      setCurrentBankAccount(newSelectedBankAccount.bankAccountId);
      return;
    }
    console.warn('no bank account could be selected, something is wrong');
  }

  const InstitutionLogo = () => {
    if (!institution?.logo) return <AccountBalance color='info' />;

    return (
      <img
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
    <div className='w-full h-12 flex items-center justify-center relative group'>
      <div className={ classes } />
      <div
        className='cursor-pointer absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle drop-shadow-md flex justify-center items-center'
        onClick={ onClick }
      >
        <InstitutionLogo />
      </div>
    </div>
  );
}
