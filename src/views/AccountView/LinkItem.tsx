import React, { Component, Fragment } from "react";
import { Map } from "immutable";
import BankAccount from "data/BankAccount";
import Link, { LinkStatus } from "data/Link";
import Balance from "data/Balance";
import { Divider, IconButton, ListItem, ListSubheader, Menu, MenuItem, Tooltip, Typography } from "@material-ui/core";
import { Autorenew, CloudOff, Edit, FiberManualRecord, MoreVert, Remove } from "@material-ui/icons";
import PlaidIcon from "Plaid/PlaidIcon";
import { getBankAccountsByLinkId } from "shared/bankAccounts/selectors/getBankAccountsByLinkId";
import { connect } from "react-redux";
import { getBalances } from "shared/balances/selectors/getBalances";
import RemoveLinkConfirmationDialog from "views/AccountView/RemoveLinkConfirmationDialog";
import { UpdatePlaidAccountDialog } from "views/AccountView/UpdatePlaidAccountDialog";

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
            <FiberManualRecord className="text-green-500 mr-2"/>
          </Tooltip>
        );
      case LinkStatus.Pending:
        return (
          <Tooltip title="This link has not been completely setup yet.">
            <FiberManualRecord className="text-yellow-500 mr-2"/>
          </Tooltip>
        );
      case LinkStatus.Error:
        return (
          <Tooltip title={ link.getErrorMessage() }>
            <FiberManualRecord className="text-red-500 mr-2"/>
          </Tooltip>
        );
      case LinkStatus.Unknown:
        return <FiberManualRecord className="text-gray-500 mr-2"/>;
    }
  };

  renderPlaidInfo = (): React.ReactNode => {
    const { link } = this.props;

    return (
      <div className="flex items-center">
        { this.renderPlaidStatus() }
        <Typography className="pr-5 items-center self-center">
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
      <Fragment>
        <ListItem key={ bankAccountId } button>
          <div className="flex w-full">
            <Typography className="w-1/3 font-bold overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap">
              { bankAccount.name }
            </Typography>
            <div className="flex-auto flex">
              <Typography className="w-1/2 m-w-1/2 overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap">
                <span
                  className="font-semibold">Safe-To-Spend:</span> { balances ? balances.getSafeToSpendString() : '...' }
              </Typography>
              <div className="w-1/2 flex">
                <Typography className="w-1/2 text-sm  overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap">
                  <span className="font-semibold">Available:</span> { bankAccount.getAvailableBalanceString() }
                </Typography>
                <Typography className="w-1/2 text-sm  overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap">
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

  renderDialogs = () => {
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
            <ListSubheader className="pl-0 pr-2 pt-2 bg-gray-50">
              <div className="flex pb-2">
                <div className="flex-auto items-center self-center">
                  <Typography className="ml-6 font-semibold text-xl h-full">
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
                .sortBy(item => item.name)
                .map(item => this.renderBankAccountItem(item.bankAccountId))
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
