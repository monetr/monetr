import React, { Component, Fragment } from "react";
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";
import { connect } from "react-redux";
import { Button, Card, Divider, Fab, List, ListItem, ListSubheader, Typography } from "@material-ui/core";
import BankAccount from "data/BankAccount";
import { Map } from 'immutable';
import { AccountBalance, Add, FiberManualRecord } from "@material-ui/icons";
import AddBankAccountDialog from "views/AccountView/AddBankAccountDialog";
import Link, { LinkStatus } from "data/Link";
import { getLinks } from "shared/links/selectors/getLinks";
import PlaidIcon from "Plaid/PlaidIcon";
import Balance from "data/Balance";
import { getBalances } from "shared/balances/selectors/getBalances";
import fetchMissingBankAccountBalances from "shared/balances/actions/fetchMissingBankAccountBalances";

interface WithConnectionPropTypes {
  bankAccounts: Map<number, BankAccount>;
  links: Map<number, Link>;
  balances: Map<number, Balance>;
  fetchMissingBankAccountBalances: () => Promise<any>;
}

enum DialogOpen {
  CreateBankAccount,
}

interface State {
  dialog: DialogOpen | null;
}

class AllAccountsView extends Component<WithConnectionPropTypes, State> {

  state = {
    dialog: null,
  };

  componentDidMount() {
    this.props.fetchMissingBankAccountBalances().then(r => {
    });
  }

  renderContents = () => {
    const { bankAccounts } = this.props;
    if (bankAccounts.isEmpty()) {
      return this.renderNoBankAccounts();
    }

    return this.renderBankAccountList();
  };

  renderNoBankAccounts = () => (
    <div className="h-full flex justify-center items-center">
      <div className="grid grid-cols-1 grid-rows-3 grid-flow-col gap-2">
        <AccountBalance className="h-32 w-full self-center opacity-40"/>
        <div className="flex items-center">
          <Typography
            className="opacity-50 text-center"
            variant="h3"
          >
            You don't have any bank accounts yet...
          </Typography>
        </div>
        <div className="w-full">
          <Button
            onClick={ this.openDialog(DialogOpen.CreateBankAccount) }
            color="primary"
            className="w-full"
          >
            <Typography
              variant="h6"
            >
              Create or Add a bank account
            </Typography>
          </Button>
        </div>
      </div>
    </div>
  );

  renderBankAccountList = () => {
    const { bankAccounts, links } = this.props;

    return (
      <Fragment>
        <List disablePadding>
          { bankAccounts
            .groupBy(item => item.linkId)
            .sortBy((_, linkId) => links.get(linkId).getName())
            .map((accounts, group) => (
              <li key={ group }>
                <ul>
                  <ListSubheader className="pl-0 pr-0 pt-2 bg-gray-50">
                    <div className="flex pb-2">
                      <div className="flex-auto">
                        <Typography className="ml-6 font-semibold text-base">
                          { links.get(group).getName() }
                        </Typography>
                      </div>
                      { links.get(group).getIsPlaid() && this.renderPlaidInfo(links.get(group)) }
                    </div>
                    <Divider/>
                  </ListSubheader>
                  {
                    accounts
                      .sortBy(item => item.name)
                      .map(item => this.renderBankAccountItem(item.bankAccountId))
                      .valueSeq()
                      .toArray()
                  }
                </ul>
              </li>
            ))
            .valueSeq()
            .toArray()
          }
        </List>
        <Fab
          color="primary"
          aria-label="add"
          className="absolute bottom-16 right-16 z-50"
          onClick={ this.openDialog(DialogOpen.CreateBankAccount) }
        >
          <Add/>
        </Fab>
      </Fragment>
    );
  };

  renderPlaidInfo = (link: Link) => {
    return (
      <div className="flex items-center">
        { link.linkStatus === LinkStatus.Setup && <FiberManualRecord className="text-green-500 mr-2"/> }
        { link.linkStatus === LinkStatus.Error && <FiberManualRecord className="text-red-500 mr-2"/> }
        <Typography className="pr-5 items-center self-center">
          <span
            className="font-bold">Last Successful Sync:</span> { link.lastSuccessfulUpdate.format('MMMM Do, h:mm a') }
        </Typography>
        <PlaidIcon className={ 'w-16 flex-none mr-6' }/>
      </div>
    )
  }

  renderBankAccountItem = (bankAccountId: number) => {
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
  }

  openDialog = (dialog: DialogOpen) => () => this.setState({
    dialog,
  });

  closeDialog = () => this.setState({ dialog: null });

  renderDialogs = () => {
    const { dialog } = this.state;

    switch (dialog) {
      case DialogOpen.CreateBankAccount:
        return <AddBankAccountDialog open={ true } onClose={ this.closeDialog }/>;
      default:
        return null;
    }
  };

  render() {
    return (
      <Fragment>
        { this.renderDialogs() }
        <div className="minus-nav">
          <div className="flex flex-col h-full md:p-10 sm:p-1 max-h-full">
            <Card elevation={ 4 } className="w-full h-full overflow-y-auto">
              { this.renderContents() }
            </Card>
          </div>
        </div>
      </Fragment>
    )
  }
}

export default connect(
  state => ({
    bankAccounts: getBankAccounts(state),
    links: getLinks(state),
    balances: getBalances(state),
  }),
  {
    fetchMissingBankAccountBalances,
  }
)(AllAccountsView);
