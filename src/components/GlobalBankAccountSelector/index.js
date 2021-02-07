import React, { Component } from 'react';
import PropTypes from "prop-types";
import { Map } from 'immutable';

export class GlobalBankAccountSelector extends Component {

  static propTypes = {
    links: PropTypes.instanceOf(Map).isRequired,
    bankAccounts: PropTypes.instanceOf(Map).isRequired,
  };

  render() {

  }
}
