import React, { Component, Fragment } from "react";
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";
import { connect } from "react-redux";
import { Button, Card, Divider, Fab, List, ListItem, ListSubheader, Typography } from "@material-ui/core";
import BankAccount from "data/BankAccount";
import { Map } from 'immutable';
import { AccountBalance, Add } from "@material-ui/icons";
import AddBankAccountDialog from "views/AccountView/AddBankAccountDialog";
import Link from "data/Link";
import { getLinks } from "shared/links/selectors/getLinks";
import PlaidIcon from "Plaid/PlaidIcon";

interface WithConnectionPropTypes {
  bankAccounts: Map<number, BankAccount>;
  links: Map<number, Link>;
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
            .map((accounts, group) => (
              <li key={ group }>
                <ul>
                  <ListSubheader className="bg-white pl-0 pr-0 pt-2 bg-gray-50">
                    <div className="flex pb-2">
                      <div className="flex-auto">
                        <Typography className="ml-6 font-semibold text-base">
                          { links.get(group).getName() }
                        </Typography>
                      </div>
                      { links.get(group).getIsPlaid() && <PlaidIcon className={ 'w-16 flex-none mr-6'} /> }
                    </div>
                    <Divider/>
                  </ListSubheader>
                  { accounts.map(item => this.renderBankAccountItem(item.bankAccountId)).valueSeq().toArray() }
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
          className="absolute bottom-16 right-16"
          onClick={ this.openDialog(DialogOpen.CreateBankAccount) }
        >
          <Add/>
        </Fab>
      </Fragment>
    );
  };

  renderBankAccountItem = (bankAccountId: number) => {
    const bankAccount = this.props.bankAccounts.get(bankAccountId);

    return (
      <ListItem key={ bankAccountId } button>
        { bankAccount.name }
      </ListItem>
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
            <Card elevation={ 4 } className="w-full h-full">
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
  }),
  {}
)(AllAccountsView);
