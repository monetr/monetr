import React, { Fragment, useState } from 'react';
import { Autorenew, CloudOff, Edit, MoreVert, PriceChange, Remove } from '@mui/icons-material';
import { Divider, IconButton, ListSubheader, Menu, MenuItem } from '@mui/material';
import * as R from 'ramda';

import LinkStatusIndicator from './LinkStatusIndicator';
import PlaidLinkStatusIndicator from './PlaidLinkStatusIndicator';

import BankLogo from '../BankLogo';

import LinkedBankAccountItem from 'components/BankAccounts/AllAccountsView/LinkedBankAccountItem';
import { showRemoveLinkConfirmationDialog } from 'components/BankAccounts/AllAccountsView/RemoveLinkConfirmationDialog';
import { showUpdatePlaidAccountDialog } from 'components/BankAccounts/AllAccountsView/UpdatePlaidAccountDialog';
import PlaidIcon from 'components/Plaid/PlaidIcon';
import { useTriggerManualSync } from 'hooks/links';
import BankAccount from 'models/BankAccount';
import Link, { LinkStatus } from 'models/Link';

interface LinkedAccountItemProps {
  link: Link;
  bankAccounts: Array<BankAccount>;
}

export default function LinkedAccountItem(props: LinkedAccountItemProps): JSX.Element {
  const triggerManualSync = useTriggerManualSync();

  async function doManualSync() {
    return triggerManualSync(props.link.linkId);
  }

  const [menuAnchor, setMenuAnchor] = useState<Element | null>();

  const openMenu = (event: { currentTarget: Element }) => setMenuAnchor(event.currentTarget);

  const closeMenu = () => setMenuAnchor(null);

  function PlaidStatus(): JSX.Element {
    if (props.link.getIsPlaid()) {
      return <PlaidLinkStatusIndicator link={ props.link } />;
    }

    return <LinkStatusIndicator link={ props.link } />;
  }

  function PlaidInfoMaybe(): JSX.Element {
    if (!props.link.getIsPlaid()) {
      return null;
    }

    return (
      <div className="flex items-center">
        <PlaidStatus />
        <PlaidIcon className="w-16 flex-none mr-6" />
      </div>
    );
  }

  const items = R.pipe(
    R.sortBy((item: BankAccount) => item.name),
    R.map((item: BankAccount) => (<LinkedBankAccountItem key={ item.bankAccountId } bankAccount={ item } />))
  )(props.bankAccounts);

  return (
    <Fragment>
      <li>
        <ul>
          <ListSubheader className="pt-2 pl-0 pr-2 bg-transparent">
            <div className="flex pb-2">
              <div className="flex flex-row self-center flex-auto h-full items-center pl-2.5">
                <BankLogo plaidInstitutionId={ props.link.plaidInstitutionId } />
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
                    onClick={ () => showUpdatePlaidAccountDialog({
                      linkId: props.link.linkId,
                      updateAccountSelection: false,
                    }) }
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
                { props.link.getIsPlaid() &&
                  <MenuItem onClick={ doManualSync }>
                    <Autorenew className="mr-2" />
                    Manually Resync
                  </MenuItem>
                }
                { props.link.getIsPlaid() &&
                  <MenuItem onClick={ () => showUpdatePlaidAccountDialog({
                    linkId: props.link.linkId,
                    updateAccountSelection: true,
                  }) }>
                    <PriceChange className="mr-2" />
                    Update Account Selection
                  </MenuItem>
                }
                <MenuItem>
                  <Edit className="mr-2" />
                  Rename
                </MenuItem>
                <MenuItem
                  className="text-red-500"
                  onClick={ () => showRemoveLinkConfirmationDialog({
                    linkId: props.link.linkId,
                  }) }
                >
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
