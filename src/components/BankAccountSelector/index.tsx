import { InputLabel, MenuItem, Select } from "@material-ui/core";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import setSelectedBankAccountId from 'shared/bankAccounts/actions/setSelectedBankAccountId';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import BankAccount from 'data/BankAccount';
import { Map } from 'immutable';
import Link from 'data/Link';

interface PropTypes {
  selectedBankAccountId: number;
  setSelectedBankAccountId: {
    (bankAccountId: number): void
  };
  bankAccounts: Map<number, BankAccount>;
  links: Map<number, Link>;
}

export class BankAccountSelector extends Component<PropTypes, {}> {

  render() {
    return (
      <Fragment>
        <InputLabel id="bank-account-selection-label">Bank Account</InputLabel>
        <Select
          labelId="bank-account-selection-label"
          id="bank-account-selection-select"
          value={ this.props.selectedBankAccountId }
          onChange={ (value) => {
            this.props.setSelectedBankAccountId(0);
          } }
          label="Bank Account"
        >
          <MenuItem value="">
            <em>None</em>
          </MenuItem>
          {
            this.props.bankAccounts.map(bankAccount => (
              <MenuItem
                value={ bankAccount.bankAccountId }
              >
                { /* make it so its the link name - bank name */ }
                { this.props.links.get(bankAccount.linkId).getName() } - { bankAccount.name }
              </MenuItem>
            ))
          }
        </Select>
      </Fragment>
    )
  }
}

export default connect(
  state => ({
    selectedBankAccountId: getSelectedBankAccountId(state),
  }),
  {
    setSelectedBankAccountId,
  },
)(BankAccountSelector);
