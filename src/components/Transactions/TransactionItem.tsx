import React, { Component } from 'react';

interface PropTypes {
  transactionId: number;
}

export class TransactionsView extends Component<PropTypes, {}> {

  render() {
    return (
      <div>
        <span>Transaction</span>
      </div>
    )
  }
}

