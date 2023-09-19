/* eslint-disable max-len */
import React, { Fragment, useCallback, useState } from 'react';
import { AutorenewOutlined, DeleteOutline, MoreVert, PriceChangeOutlined } from '@mui/icons-material';
import { Popover } from '@mui/material';

import MDivider from 'components/MDivider';
import MSpan from 'components/MSpan';
import { useSelectedBankAccount } from 'hooks/bankAccounts';
import { useLink } from 'hooks/links';
import { showRemoveLinkModal } from 'modals/RemoveLinkModal';


export default function BudgetingSidebarTitle(): JSX.Element {
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const { data: link } = useLink(bankAccount?.linkId);

  const [anchorEl, setAnchorEl] = useState<HTMLDivElement | null>(null);
  const open = Boolean(anchorEl);

  const handleClick = useCallback((event: React.MouseEvent<HTMLDivElement>) => {
    setAnchorEl(event.currentTarget);
  }, [setAnchorEl]);
  const handleClose = useCallback(() => setAnchorEl(null), [setAnchorEl]);

  const handleRemoveLink = useCallback(() => {
    setAnchorEl(null);
    showRemoveLinkModal({ link: link });
  }, [setAnchorEl, link]);

  return (
    <Fragment>
      <div
        onClick={ handleClick }
        className='flex h-12 w-full items-center p-2 dark:text-dark-monetr-content-emphasis dark:hover:bg-dark-monetr-background-emphasis cursor-pointer'
      >
        <span className='truncate text-xl font-semibold'>
          {link?.getName()}
        </span>
        <MoreVert className='ml-auto' />
      </div>
      <MDivider className='w-1/2' />
      <Popover
        open={ open }
        anchorEl={ anchorEl }
        onClose={ handleClose }
        transitionDuration={ 100 }
        anchorOrigin={ {
          vertical: 'bottom',
          horizontal: 'left',
        } }
        className='ml-[5px]'
      >
        <div className='flex flex-col dark:bg-dark-monetr-background rounded-lg border dark:border-dark-monetr-border-subtle dark:shadow-2xl' style={ { width: `${anchorEl?.offsetWidth - 10}px` } }>
          <MSpan
            size='lg'
            weight='medium'
            className='p-2 cursor-pointer dark:hover:bg-dark-monetr-background-emphasis dark:hover:text-dark-monetr-content-emphasis'
            component='a'
            ellipsis
          >
            <PriceChangeOutlined className='mr-1 mb-0.5' />
            Update Account Selection
          </MSpan>
          <MSpan
            size='lg'
            weight='medium'
            className='p-2 cursor-pointer dark:hover:bg-dark-monetr-background-emphasis dark:hover:text-dark-monetr-content-emphasis'
            component='a'
            ellipsis
          >
            <AutorenewOutlined className='mr-1 mb-0.5' />
            Manually Resync
          </MSpan>
          <MDivider />
          <MSpan
            size='lg'
            weight='medium'
            className='p-2 cursor-pointer dark:hover:bg-dark-monetr-background-emphasis dark:hover:text-dark-monetr-content-emphasis'
            component='a'
            ellipsis
            onClick={ handleRemoveLink }
          >
            <DeleteOutline className='mr-1 mb-0.5 dark:text-dark-monetr-red' />
            Remove { link?.getName() }
          </MSpan>
        </div>
      </Popover>
    </Fragment>
  );
}
