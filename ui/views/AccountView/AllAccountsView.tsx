import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { AppState } from 'store';

import AddBankAccountDialog from 'views/AccountView/AddBankAccountDialog';
import Balance from 'models/Balance';
import BankAccount from 'models/BankAccount';
import Link from 'models/Link';
import LinkItem from 'views/AccountView/LinkItem';
import fetchMissingBankAccountBalances from 'shared/balances/actions/fetchMissingBankAccountBalances';
import { AccountBalance, Add } from '@mui/icons-material';
import { Button, Card, Fab, List, Typography } from '@mui/material';
import { Map } from 'immutable';
import { getBalances } from 'shared/balances/selectors/getBalances';
import { getBankAccounts } from 'shared/bankAccounts/selectors/getBankAccounts';
import { getLinks } from 'shared/links/selectors/getLinks';

interface WithConnectionPropTypes {
  bankAccounts: Map<number, BankAccount>;
  links: Map<number, Link>;
  balances: Map<number, Balance>;
  fetchMissingBankAccountBalances: () => Promise<any>;
}

enum DialogOpen {
  CreateBankAccount,
}

interface ComponentState {
  dialog: DialogOpen | null;
  menuAnchorEl: Element | null;
}

class AllAccountsView extends Component<WithConnectionPropTypes, ComponentState> {

  state = {
    dialog: null,
    menuAnchorEl: null,
  };

  componentDidMount() {
    this.props.fetchMissingBankAccountBalances().catch(error => {
      console.error(error);
    });
  }

  renderContents = (): React.ReactNode => {
    const { bankAccounts } = this.props;
    if (bankAccounts.isEmpty()) {
      return this.renderNoBankAccounts();
    }

    return this.renderBankAccountList();
  };

  renderNoBankAccounts = (): React.ReactNode => (
    <div className="flex items-center justify-center h-full">
      <div className="grid grid-cols-1 grid-rows-3 grid-flow-col gap-2">
        <AccountBalance className="self-center w-full h-32 opacity-40"/>
        <div className="flex items-center">
          <Typography
            className="text-center opacity-50"
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

  renderBankAccountList = (): React.ReactNode => {
    const { bankAccounts, links } = this.props;

    return (
      <Fragment>
        <List disablePadding>
          { bankAccounts
            .groupBy((item: BankAccount) => item.linkId)
            .sortBy((_, linkId: number) => links.get(linkId).getName())
            .map((_, linkId: number) => (
              <LinkItem
                key={ linkId }
                link={ links.get(linkId) }
              />
            ))
            .valueSeq()
            .toArray()
          }
        </List>
        <Fab
          color="primary"
          aria-label="add"
          className="absolute z-50 bottom-0 right-5"
          onClick={ this.openDialog(DialogOpen.CreateBankAccount) }
        >
          <Add/>
        </Fab>
      </Fragment>
    );
  };

  openDialog = (dialog: DialogOpen) => () => this.setState({
    dialog,
  });

  closeDialog = () => this.setState({ dialog: null });

  renderDialogs = (): React.ReactNode | null => {
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
          <div className="flex flex-col h-full max-h-full">
            { this.renderContents() }
          </div>
        </div>
      </Fragment>
    )
  }
}

export default connect(
  (state: AppState) => ({
    bankAccounts: getBankAccounts(state),
    links: getLinks(state),
    balances: getBalances(state),
  }),
  {
    fetchMissingBankAccountBalances,
  }
)(AllAccountsView);
