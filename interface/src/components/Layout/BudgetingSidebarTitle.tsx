/* eslint-disable max-len */
import React, { Fragment, useCallback } from 'react';
import { EllipsisVertical, LogIn, Plug, RefreshCw, Trash2 } from 'lucide-react';

import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@monetr/interface/components/DropdownMenu';
import MDivider from '@monetr/interface/components/MDivider';
import { ReactElement } from '@monetr/interface/components/types';
import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useLink, useTriggerManualPlaidSync } from '@monetr/interface/hooks/links';
import { showRemoveLinkModal } from '@monetr/interface/modals/RemoveLinkModal';
import { showUpdatePlaidAccountOverlay } from '@monetr/interface/modals/UpdatePlaidAccountOverlay';

export default function BudgetingSidebarTitle(): JSX.Element {
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: link } = useLink(bankAccount?.linkId);
  const triggerSync = useTriggerManualPlaidSync();

  const handleReauthenticateLink = useCallback(() => {
    showUpdatePlaidAccountOverlay({
      link: link,
    });
  }, [link]);

  const handleTriggerResync = useCallback(() => {
    triggerSync(bankAccount?.linkId);
  }, [bankAccount?.linkId, triggerSync]);

  const handleUpdateAccountSelection = useCallback(() => {
    showUpdatePlaidAccountOverlay({
      link: link,
      updateAccountSelection: true,
    });
  }, [link]);

  const handleRemoveLink = useCallback(() => {
    showRemoveLinkModal({ link: link });
  }, [link]);

  if (!link) {
    return null;
  }

  return (
    <Fragment>
      <DropdownMenu>
        <DropdownMenuTrigger
          className='flex h-12 w-full items-center p-2 dark:text-dark-monetr-content-emphasis dark:hover:bg-dark-monetr-background-emphasis'
        >
          <span className='truncate text-xl font-semibold'>
            { link?.getName() }
          </span>
          <EllipsisVertical className='ml-auto shrink-0' />
        </DropdownMenuTrigger>
        <DropdownMenuContent className='w-72'>
          <MenuItem 
            visible={ link.getIsPlaid() && (link.getIsError() || link.getIsPendingExpiration()) } 
            onClick={ handleReauthenticateLink }
          >
            <LogIn />
            Reauthenticate
          </MenuItem>
          <MenuItem visible={ link.getIsPlaid() } onClick={ handleUpdateAccountSelection }>
            <Plug />
            Update Account Selection
          </MenuItem>
          <MenuItem visible={ link.getIsPlaid() } onClick={ handleTriggerResync }>
            <RefreshCw />
            Manually Resync
          </MenuItem>
          <MDivider />
          <MenuItem visible onClick={ handleRemoveLink }>
            <Trash2 className='text-dark-monetr-red' />
            Remove { link?.getName() }
          </MenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <MDivider className='w-1/2' />
    </Fragment>
  );
}

interface MenuItemProps {
  visible?: boolean;
  onClick: () => unknown;
  children?: ReactElement;
}

function MenuItem({ visible, onClick, children }: MenuItemProps): JSX.Element {
  if (!visible) {
    return null;
  }

  return (
    <DropdownMenuItem
      onClick={ onClick }
    >
      { children }
    </DropdownMenuItem>
  );
}
