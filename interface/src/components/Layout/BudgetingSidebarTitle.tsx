import { Fragment, useCallback } from 'react';
import { EllipsisVertical, LogIn, Plug, RefreshCw, Settings, Trash2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

import Divider from '@monetr/interface/components/Divider';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@monetr/interface/components/DropdownMenu';
import { Skeleton } from '@monetr/interface/components/Skeleton';
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
        <Divider className='w-1/2' />
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
            onClick={handleReauthenticateLink}
            visible={link.getIsPlaid() && (link.getIsError() || link.getIsPendingExpiration())}
          >
            <LogIn />
            Reauthenticate
          </MenuItem>
          <MenuItem onClick={handleUpdateAccountSelection} visible={link.getIsPlaid()}>
            <Plug />
            Update Account Selection
          </MenuItem>
          <MenuItem onClick={handleTriggerResync} visible={link.getIsPlaid() && !link.getIsRevoked()}>
            <RefreshCw />
            Manually Resync
          </MenuItem>
          <MenuItem onClick={handleLinkSettings} visible>
            <Settings />
            Settings
          </MenuItem>
          <Divider />
          <MenuItem onClick={handleRemoveLink} visible>
            <Trash2 className='text-dark-monetr-red' />
            Remove {link?.getName()}
          </MenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <Divider className='w-1/2' />
    </Fragment>
  );
}

interface MenuItemProps {
  visible?: boolean;
  onClick: () => unknown;
  children?: React.ReactNode;
}

function MenuItem({ visible, onClick, children }: MenuItemProps): JSX.Element {
  if (!visible) {
    return null;
  }

  return <DropdownMenuItem onClick={onClick}>{children}</DropdownMenuItem>;
}
