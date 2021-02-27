import { InputLabel, MenuItem, Select } from "@material-ui/core";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import setSelectedBankAccountId from 'shared/bankAccounts/actions/setSelectedBankAccountId';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import BankAccount from 'data/BankAccount';
import { Map } from 'immutable';
import Link from 'data/Link';
import { getLinks } from "shared/links/selectors/getLinks";
import { getBankAccounts } from "shared/bankAccounts/selectors/getBankAccounts";
import { getBankAccountsLoading } from "shared/bankAccounts/selectors/getBankAccountsLoading";
import { getLinksLoading } from "shared/links/selectors/getLinksLoading";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";

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
}

export class BankAccountSelector extends Component<PropTypes, {}> {

  changeBankAccount = (event) => {
    this.props.setSelectedBankAccountId(event.target.value as number);
    this.props.fetchInitialTransactionsIfNeeded();
  };

  render() {
    const { bankAccountsLoading, linksLoading } = this.props;

    if (bankAccountsLoading || linksLoading) {
      return null;
    }

    return (
      <Fragment>
        <InputLabel id="bank-account-selection-label">Bank Account</InputLabel>
        <Select
          labelId="bank-account-selection-label"
          id="bank-account-selection-select"
          value={ this.props.selectedBankAccountId || this.props.bankAccounts.first<BankAccount>().bankAccountId }
          onChange={ this.changeBankAccount }
          label="Bank Account"
        >
          {
            this.props.bankAccounts.map(bankAccount => {
              const link = this.props.links.get(bankAccount.linkId);
              return (
                <MenuItem
                  value={ bankAccount.bankAccountId }
                >
                  { /* make it so its the link name - bank name */ }
                  { link.getName() } - { bankAccount.name }
                </MenuItem>
              )
            })
          }
        </Select>
      </Fragment>
    )
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
  },
)(BankAccountSelector);
