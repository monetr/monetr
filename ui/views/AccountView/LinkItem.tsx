import React, { Component, Fragment } from 'react';
import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import Link, { LinkStatus } from 'models/Link';
import Balance from 'models/Balance';
import { Divider, IconButton, ListItem, ListSubheader, Menu, MenuItem, Tooltip, Typography } from '@mui/material';
import { Autorenew, CloudOff, Edit, FiberManualRecord, MoreVert, Remove } from '@mui/icons-material';
import PlaidIcon from 'components/Plaid/PlaidIcon';
import { getBankAccountsByLinkId } from 'shared/bankAccounts/selectors/getBankAccountsByLinkId';
import { connect } from 'react-redux';
import { getBalances } from 'shared/balances/selectors/getBalances';
import RemoveLinkConfirmationDialog from 'views/AccountView/RemoveLinkConfirmationDialog';
import { UpdatePlaidAccountDialog } from 'views/AccountView/UpdatePlaidAccountDialog';

interface PropTypes {
  link: Link;
}

interface WithConnectionPropTypes extends PropTypes {
  balances: Map<number, Balance>;
  bankAccounts: Map<number, BankAccount>;
}

enum DialogOpen {
  RemoveLinkDialog,
  UpdateLinkDialog,
}

interface State {
  dialog: DialogOpen | null;
  menuAnchorEl: Element | null;
}

class LinkItem extends Component<WithConnectionPropTypes, State> {

  state = {
    dialog: null,
    menuAnchorEl: null,
  };

  renderPlaidStatus = (): React.ReactNode => {
    const { link } = this.props;

    switch (link.linkStatus) {
      case LinkStatus.Setup:
        return (
          <Tooltip title="This link is working properly.">
            <FiberManualRecord className="mr-2 text-green-500"/>
          </Tooltip>
        );
      case LinkStatus.Pending:
        return (
          <Tooltip title="This link has not been completely setup yet.">
            <FiberManualRecord className="mr-2 text-yellow-500"/>
          </Tooltip>
        );
      case LinkStatus.Error:
        return (
          <Tooltip title={ link.getErrorMessage() }>
            <FiberManualRecord className="mr-2 text-red-500"/>
          </Tooltip>
        );
      case LinkStatus.Unknown:
        return <FiberManualRecord className="mr-2 text-gray-500"/>;
    }
  };

  renderPlaidInfo = (): React.ReactNode => {
    const { link } = this.props;

    return (
      <div className="flex items-center">
        { this.renderPlaidStatus() }
        <Typography className="items-center self-center pr-5">
          <span
            className="font-bold">Last Successful Sync:</span> { link.lastSuccessfulUpdate ? link.lastSuccessfulUpdate.format('MMMM Do, h:mm a') : 'N/A' }
        </Typography>
        <PlaidIcon className={ 'w-16 flex-none mr-6' }/>
      </div>
    );
  };

  renderBankAccountItem = (bankAccountId: number): React.ReactNode => {
    const bankAccount = this.props.bankAccounts.get(bankAccountId);
    const balances = this.props.balances.get(bankAccountId, null);

    return (
      <Fragment key={ bankAccountId }>
        <ListItem button>
          <div className="flex w-full">
            <Typography className="w-1/3 overflow-hidden font-bold overflow-ellipsis flex-nowrap whitespace-nowrap">
              { bankAccount.name }
            </Typography>
            <div className="flex flex-auto">
              <Typography className="w-1/2 overflow-hidden m-w-1/2 overflow-ellipsis flex-nowrap whitespace-nowrap">
                <span
                  className="font-semibold">Safe-To-Spend:</span> { balances ? balances.getSafeToSpendString() : '...' }
              </Typography>
              <div className="flex w-1/2">
                <Typography className="w-1/2 overflow-hidden text-sm overflow-ellipsis flex-nowrap whitespace-nowrap">
                  <span className="font-semibold">Available:</span> { bankAccount.getAvailableBalanceString() }
                </Typography>
                <Typography className="w-1/2 overflow-hidden text-sm overflow-ellipsis flex-nowrap whitespace-nowrap">
                  <span className="font-semibold">Current:</span> { bankAccount.getCurrentBalanceString() }
                </Typography>
              </div>
            </div>
          </div>
        </ListItem>
        <Divider/>
      </Fragment>
    )
  };

  openMenu = (event: { currentTarget: Element }) => this.setState({
    menuAnchorEl: event.currentTarget,
  });

  closeMenu = () => this.setState({
    menuAnchorEl: null,
  });

  openDialog = (dialog: DialogOpen) => () => this.setState({
    dialog,
    menuAnchorEl: null,
  });

  closeDialog = () => this.setState({ dialog: null });

  renderDialogs = (): React.ReactNode | null => {
    const { dialog } = this.state;

    switch (dialog) {
      case DialogOpen.RemoveLinkDialog:
        return <RemoveLinkConfirmationDialog
          open
          onClose={ this.closeDialog }
          linkId={ this.props.link.linkId }
        />;
      case DialogOpen.UpdateLinkDialog:
        return <UpdatePlaidAccountDialog open onClose={ this.closeDialog } linkId={ this.props.link.linkId }/>;
      default:
        return null;
    }
  };

  render() {
    const { link, bankAccounts } = this.props;

    return (
      <Fragment>
        { this.renderDialogs() }

        <li>
          <ul>
            <ListSubheader className="pt-2 pl-0 pr-2">
              <div className="flex pb-2">
                <div className="items-center self-center flex-auto">
                  <Typography className="h-full ml-6 text-xl font-semibold">
                    { link.getName() }
                  </Typography>
                </div>
                { link.getIsPlaid() && this.renderPlaidInfo() }
                <IconButton onClick={ this.openMenu }>
                  <MoreVert/>
                </IconButton>
                <Menu
                  id={ `${ link.linkId }-link-menu` }
                  anchorEl={ this.state.menuAnchorEl }
                  keepMounted
                  open={ Boolean(this.state.menuAnchorEl) }
                  onClose={ this.closeMenu }
                >
                  { link.getIsPlaid() && link.linkStatus === LinkStatus.Error &&
                  <MenuItem
                    onClick={ this.openDialog(DialogOpen.UpdateLinkDialog) }
                    className="text-yellow-600"
                  >
                    <Autorenew className="mr-2"/>
                    Reauthenticate
                  </MenuItem>
                  }
                  { link.getIsPlaid() &&
                  <MenuItem>
                    <CloudOff className="mr-2"/>
                    Convert To Manual Link
                  </MenuItem>
                  }
                  <MenuItem>
                    <Edit className="mr-2"/>
                    Rename
                  </MenuItem>
                  <MenuItem className="text-red-500" onClick={ this.openDialog(DialogOpen.RemoveLinkDialog) }>
                    <Remove className="mr-2"/>
                    Remove
                  </MenuItem>
                </Menu>
              </div>
              <Divider/>
            </ListSubheader>
            {
              bankAccounts
                .sortBy((item: BankAccount) => item.name)
                .map((item: BankAccount) => this.renderBankAccountItem(item.bankAccountId))
                .valueSeq()
                .toArray()
            }
          </ul>
        </li>
      </Fragment>
    );
  }
}

export default connect(
  (state, props: PropTypes) => ({
    bankAccounts: getBankAccountsByLinkId(props.link.linkId)(state),
    balances: getBalances(state),
  }),
  {}
)(LinkItem);
