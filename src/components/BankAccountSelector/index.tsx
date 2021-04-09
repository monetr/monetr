import { InputLabel, MenuItem, Select } from "@material-ui/core";
import BankAccount from 'data/BankAccount';
import Link from 'data/Link';
import { Map } from 'immutable';
import React, { ChangeEvent, Component, Fragment } from 'react';
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
  value: number;
}

export class BankAccountSelector extends Component<PropTypes, {}> {

  changeBankAccount = (event: ChangeEvent<SelectEvent>) => {
    this.props.setSelectedBankAccountId(event.target.value as number);
    return Promise.all([
      this.props.fetchInitialTransactionsIfNeeded(),
      this.props.fetchFundingSchedulesIfNeeded(),
      this.props.fetchSpending(),
      this.props.fetchBalances(),
    ])
  };

  render() {
    const { bankAccountsLoading, linksLoading } = this.props;

    if (bankAccountsLoading || linksLoading) {
      return null;
    }

    return (
      <Fragment>
        <InputLabel id="bank-account-selection-label" className="text-gray-200">Bank Account</InputLabel>
        <Select
          labelId="bank-account-selection-label"
          id="bank-account-selection-select"
          value={ this.props.selectedBankAccountId ?? this.props.bankAccounts.first<BankAccount>()?.bankAccountId }
          onChange={ this.changeBankAccount }
          label="Bank Account"
          className="text-gray-200"
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
    fetchFundingSchedulesIfNeeded,
    fetchSpending,
    fetchBalances,
  },
)(BankAccountSelector);
