/* eslint-disable max-len */
import React, { Fragment, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { EllipsisVertical, LogIn, Plug, RefreshCw, Settings, Trash2 } from 'lucide-react';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@monetr/interface/components/DropdownMenu';
import MDivider from '@monetr/interface/components/MDivider';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import type { ReactElement } from '@monetr/interface/components/types';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useTriggerManualPlaidSync } from '@monetr/interface/hooks/useTriggerManualPlaidSync';
import { showRemoveLinkModal } from '@monetr/interface/modals/RemoveLinkModal';
import { showUpdatePlaidAccountOverlay } from '@monetr/interface/modals/UpdatePlaidAccountOverlay';

export default function BudgetingSidebarTitle(): JSX.Element {
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: link } = useCurrentLink();
  const navigate = useNavigate();
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

  const handleLinkSettings = useCallback(() => {
    navigate(`/link/${link?.linkId}/details`);
  }, [link, navigate]);

  if (!link) {
    return (
      <Fragment>
        <div className='flex h-12 w-full items-center p-2 dark:text-dark-monetr-content-emphasis'>
          <Skeleton className='h-7 w-full' />
        </div>
        <MDivider className='w-1/2' />
      </Fragment>
    );
  }

  return (
    <Fragment>
      <DropdownMenu>
        <DropdownMenuTrigger className='flex h-12 w-full items-center p-2 dark:text-dark-monetr-content-emphasis dark:hover:bg-dark-monetr-background-emphasis'>
          <span className='truncate text-xl font-semibold'>{link?.getName()}</span>
          <EllipsisVertical className='ml-auto shrink-0' />
        </DropdownMenuTrigger>
        <DropdownMenuContent className='w-72'>
          <MenuItem
            visible={link.getIsPlaid() && (link.getIsError() || link.getIsPendingExpiration())}
            onClick={handleReauthenticateLink}
          >
            <LogIn />
            Reauthenticate
          </MenuItem>
          <MenuItem visible={link.getIsPlaid()} onClick={handleUpdateAccountSelection}>
            <Plug />
            Update Account Selection
          </MenuItem>
          <MenuItem visible={link.getIsPlaid() && !link.getIsRevoked()} onClick={handleTriggerResync}>
            <RefreshCw />
            Manually Resync
          </MenuItem>
          <MenuItem visible onClick={handleLinkSettings}>
            <Settings />
            Settings
          </MenuItem>
          <MDivider />
          <MenuItem visible onClick={handleRemoveLink}>
            <Trash2 className='text-dark-monetr-red' />
            Remove {link?.getName()}
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

  return <DropdownMenuItem onClick={onClick}>{children}</DropdownMenuItem>;
}
