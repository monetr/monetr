import React, { Fragment, useState } from 'react';
import { useQuery } from 'react-query';
import { Autorenew, CloudOff, Edit, FiberManualRecord, MoreVert, Remove } from '@mui/icons-material';
import { Divider, IconButton, ListSubheader, Menu, MenuItem, Tooltip, Typography } from '@mui/material';
import * as R from 'ramda';

import LinkedBankAccountItem from 'components/BankAccounts/AllAccountsView/LinkedBankAccountItem';
import RemoveLinkConfirmationDialog from 'components/BankAccounts/AllAccountsView/RemoveLinkConfirmationDialog';
import UpdatePlaidAccountDialog from 'components/BankAccounts/AllAccountsView/UpdatePlaidAccountDialog';
import PlaidIcon from 'components/Plaid/PlaidIcon';
import BankAccount from 'models/BankAccount';
import Link, { LinkStatus } from 'models/Link';

interface LinkedAccountItemProps {
  link: Link;
  bankAccounts: Array<BankAccount>;
}

enum DialogOpen {
  RemoveLinkDialog,
  UpdateLinkDialog,
}

export default function LinkedAccountItem(props: LinkedAccountItemProps): JSX.Element {
  const { data } = useQuery<{ logo: string }>(`/institutions/${ props.link.plaidInstitutionId }`, {
    enabled: !!props.link.plaidInstitutionId,
    staleTime: 60 * 60 * 1000, // 60 minutes
  });

  const [menuAnchor, setMenuAnchor] = useState<Element | null>();
  const [dialog, setDialog] = useState<DialogOpen | null>();

  const openMenu = (event: { currentTarget: Element }) => setMenuAnchor(event.currentTarget);

  const closeMenu = () => setMenuAnchor(null);

  const openDialog = (dialog: DialogOpen) => () => {
    setDialog(dialog);
    closeMenu();
  };

  function closeDialog() {
    return setDialog(null);
  }

  function PlaidStatus(): JSX.Element {
    switch (props.link.linkStatus) {
      case LinkStatus.Setup:
        return (
          <Tooltip title="This link is working properly.">
            <FiberManualRecord className="mr-2 text-green-500" />
          </Tooltip>
        );
      case LinkStatus.Pending:
        return (
          <Tooltip title="This link has not been completely setup yet.">
            <FiberManualRecord className="mr-2 text-yellow-500" />
          </Tooltip>
        );
      case LinkStatus.Error:
        return (
          <Tooltip title={ props.link.getErrorMessage() }>
            <FiberManualRecord className="mr-2 text-red-500" />
          </Tooltip>
        );
      case LinkStatus.Unknown:
        return <FiberManualRecord className="mr-2 text-gray-500" />;
    }
  }

  function PlaidInfoMaybe(): JSX.Element {
    if (!props.link.getIsPlaid()) {
      return null;
    }

    return (
      <div className="flex items-center">
        <PlaidStatus />
        <Typography className="items-center self-center pr-5">
          <span
            className="font-bold">
            Last Successful Sync:
          </span>
          { props.link.lastSuccessfulUpdate ? props.link.lastSuccessfulUpdate.format('MMMM Do, h:mm a') : 'N/A' }
        </Typography>
        <PlaidIcon className="w-16 flex-none mr-6" />
      </div>
    );
  }

  function Dialogs(): JSX.Element {
    switch (dialog) {
      case DialogOpen.UpdateLinkDialog:
        return <UpdatePlaidAccountDialog open={ true } onClose={ closeDialog } linkId={ props.link.linkId } />;
      case DialogOpen.RemoveLinkDialog:
        return <RemoveLinkConfirmationDialog open={ true } onClose={ closeDialog } linkId={ props.link.linkId } />;
      default:
        return null;
    }
  }

  const items = R.pipe(
    R.sortBy((item: BankAccount) => item.name),
    R.map((item: BankAccount) => (<LinkedBankAccountItem key={ item.bankAccountId } bankAccount={ item } />))
  )(props.bankAccounts);

  return (
    <Fragment>
      <Dialogs />
      <li>
        <ul>
          <ListSubheader className="pt-2 pl-0 pr-2 bg-transparent">
            <div className="flex pb-2">
              <div className="flex flex-row self-center flex-auto h-full items-center pl-2.5">
                { data?.logo && <img className="max-h-8 col-span-4" src={ `data:image/png;base64,${ data.logo }` } /> }
                <span className="h-full ml-2.5 text-xl font-semibold align-middle">
                  { props.link.getName() }
                </span>
              </div>
              <PlaidInfoMaybe />
              <IconButton onClick={ openMenu }>
                <MoreVert />
              </IconButton>
              <Menu
                id={ `${ props.link.linkId }-link-menu` }
                anchorEl={ menuAnchor }
                keepMounted
                open={ Boolean(menuAnchor) }
                onClose={ closeMenu }
              >
                { props.link.getIsPlaid() && props.link.linkStatus === LinkStatus.Error &&
                  <MenuItem
                    onClick={ openDialog(DialogOpen.UpdateLinkDialog) }
                    className="text-yellow-600"
                  >
                    <Autorenew className="mr-2" />
                    Reauthenticate
                  </MenuItem>
                }
                { props.link.getIsPlaid() &&
                  <MenuItem>
                    <CloudOff className="mr-2" />
                    Convert To Manual Link
                  </MenuItem>
                }
                <MenuItem>
                  <Edit className="mr-2" />
                  Rename
                </MenuItem>
                <MenuItem className="text-red-500" onClick={ openDialog(DialogOpen.RemoveLinkDialog) }>
                  <Remove className="mr-2" />
                  Remove
                </MenuItem>
              </Menu>
            </div>
            <Divider />
          </ListSubheader>
          { items }
        </ul>
      </li>
    </Fragment>
  );
}
