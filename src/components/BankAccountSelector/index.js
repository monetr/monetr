import { InputLabel, MenuItem, Select } from "@material-ui/core";
import React, { Component, Fragment } from 'react';

export class BankAccountSelector extends Component {

  render() {
    return (
      <Fragment>
        <InputLabel id="bank-account-selection-label">Bank Account</InputLabel>
        <Select
          labelId="bank-account-selection-label"
          id="bank-account-selection-select"
          value={ 10 }
          onChange={ () => {
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
