import { InputLabel, MenuItem, Select } from "@material-ui/core";
import PropTypes from "prop-types";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import setSelectedBankAccountId from "shared/bankAccounts/actions/setSelectedBankAccountId";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";

export class BankAccountSelector extends Component {

  static propTypes = {
    selectedBankAccountId: PropTypes.number.isRequired,
    setSelectedBankAccountId: PropTypes.func
  };

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
          <MenuItem value={ 10 }>US Bank</MenuItem>
          <MenuItem value={ 20 }>Twenty</MenuItem>
          <MenuItem value={ 30 }>Thirty</MenuItem>
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
