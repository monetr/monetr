/* eslint-disable max-len */
import React from 'react';
import { AccountBalance } from '@mui/icons-material';
import clsx from 'clsx';

import { useInstitution } from 'hooks/institutions';

interface BankSidebarItemProps {
  instituionId: string;
  active?: boolean;
  onClick: () => void;
}

export default function BankSidebarItem(props: BankSidebarItemProps): JSX.Element {
  const { result: institution } = useInstitution(props.instituionId);

  const InstitutionLogo = () => {
    if (!institution?.logo) return <AccountBalance color='info' />;

    return (
      <img
        src={ `data:image/png;base64,${institution.logo}` }
      />
    );
  };

  const classes = clsx(
    'absolute',
    'dark:bg-dark-monetr-border',
    'right-0',
    'rounded-l-xl',
    'transition-transform',
    'w-1.5',
    {
      'h-8': props.active,
      'scale-y-100': props.active,
    },
    {
      'h-4': !props.active,
      'group-hover:scale-y-100': !props.active,
      'group-hover:scale-x-100': !props.active,
      'scale-x-0': !props.active,
      'scale-y-50': !props.active,
    },
  );

  return (
    <div className='w-full h-12 flex items-center justify-center relative group' onClick={ props.onClick }>
      <div className={ classes } />
      <div className='cursor-pointer absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle drop-shadow-md flex justify-center items-center'>
        <InstitutionLogo />
      </div>
    </div>
  );
}
