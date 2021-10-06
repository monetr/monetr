import { Button, Divider, Menu, MenuItem, Typography } from "@material-ui/core";
import BankAccount from 'data/BankAccount';
import Link from 'data/Link';
import { Map } from 'immutable';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import setSelectedBankAccountId from 'shared/bankAccounts/actions/setSelectedBankAccountId';
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";
import { getBankAccountsLoading } from "shared/bankAccounts/selectors/getBankAccountsLoading";
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import { getLinks } from "shared/links/selectors/getLinks";
import { getLinksLoading } from "shared/links/selectors/getLinksLoading";
import fetchSpending from 'shared/spending/actions/fetchSpending';
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import { ArrowDropDown, CheckCircle } from "@material-ui/icons";
import classnames from "classnames";
import { RouteComponentProps, withRouter } from "react-router-dom";

interface PropTypes {
  selectedBankAccountId: number;
  setSelectedBankAccountId: {
    (bankAccountId: number): void
  };
  bankAccounts: Map<number, BankAccount>;
  bankAccountsLoading: boolean;
  links: Map<number, Link>;
  linksLoading: boolean;
  fetchInitialTransactionsIfNeeded: {
    (): Promise<void>;
  };
  fetchFundingSchedulesIfNeeded: { (): Promise<void> }
  fetchSpending: { (): Promise<void> }
  fetchBalances: { (): Promise<void> }
}

interface SelectEvent {
  name: string;
  value: number | string;
}

interface State {
  anchorEl: Element | null;
}

export class BankAccountSelector extends Component<RouteComponentProps & PropTypes, State> {

  state = {
    anchorEl: null,
  };

  changeBankAccount = (bankAccountId: number) => () => {
    this.props.setSelectedBankAccountId(bankAccountId);
    this.closeMenu();
    return Promise.all([
      this.props.fetchInitialTransactionsIfNeeded(),
      this.props.fetchFundingSchedulesIfNeeded(),
      this.props.fetchSpending(),
      this.props.fetchBalances(),
    ])
  };

  openMenu = event => this.setState({
    anchorEl: event.currentTarget,
  });

  closeMenu = () => this.setState({
    anchorEl: null,
  });

  goToAllAccounts = () => {
    this.closeMenu();
    this.props.history.push('/accounts');
  }

  renderBankAccountMenu = (): JSX.Element | JSX.Element[] => {
    const { selectedBankAccountId, bankAccounts } = this.props;

    const addBankAccountItem = (
      <MenuItem key="addBankAccount">
        <Typography>
          Add Bank Account (WIP)
        </Typography>
      </MenuItem>
    );

    const bankAccountsViewButton = (
      <MenuItem key="viewBankAccounts" onClick={ this.goToAllAccounts }>
        <Typography>
          View Bank Accounts
        </Typography>
      </MenuItem>
    );

    if (bankAccounts.isEmpty()) {
      return [addBankAccountItem, bankAccountsViewButton]
    }

    let items = bankAccounts
      .sortBy(bankAccount => {
        const link = this.props.links.get(bankAccount.linkId);

        return `${ link.getName() } - ${ bankAccount.name }`;
      })
      .map(bankAccount => {
        const link = this.props.links.get(bankAccount.linkId);
        return (
          <MenuItem
            key={ bankAccount.bankAccountId }
            onClick={ this.changeBankAccount(bankAccount.bankAccountId) }
          >
            <CheckCircle color="primary" className={ classnames('mr-1', {
              'opacity-0': bankAccount.bankAccountId !== selectedBankAccountId,
            }) }/>
            { /* make it so its the link name - bank name */ }
            { link.getName() } - { bankAccount.name }
          </MenuItem>
        )
      })
      .valueSeq()
      .toArray();

    items.push(<Divider key="divider" className="w-96"/>);
    items.push(addBankAccountItem);
    items.push(bankAccountsViewButton);

    return items;
  };

  render() {
    const { bankAccountsLoading, linksLoading, selectedBankAccountId, bankAccounts } = this.props;

    if (bankAccountsLoading || linksLoading) {
      return null;
    }

    let title = "Select A Bank Account";
    if (selectedBankAccountId) {
      title = bankAccounts.get(selectedBankAccountId, null)?.name;
    }

    return (
      <Fragment>
        <Button
          className="text-white"
          onClick={ this.openMenu }
          aria-label="menu"
        >
          <Typography
            color="inherit"
            className="mr-1"
            variant="h6"
          >
            { title }
          </Typography>
          <ArrowDropDown scale={ 1.25 } color="inherit"/>
        </Button>
        <Menu
          className="w-96 pt-0 pb-0"
          id="bank-account-menu"
          anchorEl={ this.state.anchorEl }
          keepMounted
          open={ !!this.state.anchorEl }
          onClose={ this.closeMenu }
        >
          { this.renderBankAccountMenu() }
        </Menu>
      </Fragment>
    );
  }
}

export default connect(
  state => ({
    selectedBankAccountId: getSelectedBankAccountId(state),
    bankAccounts: getBankAccounts(state),
    bankAccountsLoading: getBankAccountsLoading(state),
    links: getLinks(state),
    linksLoading: getLinksLoading(state),
  }),
  {
    setSelectedBankAccountId,
    fetchInitialTransactionsIfNeeded,
    fetchFundingSchedulesIfNeeded,
    fetchSpending,
    fetchBalances,
  },
)(withRouter(BankAccountSelector));
