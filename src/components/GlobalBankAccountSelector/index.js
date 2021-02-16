import { Map } from 'immutable';
import PropTypes from "prop-types";
import React, { Component } from 'react';

export class GlobalBankAccountSelector extends Component {

  static propTypes = {
    links: PropTypes.instanceOf(Map).isRequired,
    bankAccounts: PropTypes.instanceOf(Map).isRequired,
  };

  render() {

  }
}
