import { OrderedMap } from "immutable";
import Transaction from "data/Transaction";
import React, { Component } from "react";
import { connect } from "react-redux";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";

interface PropTypes {
  transactions: OrderedMap<number, Transaction>;
  fetchInitialTransactionsIfNeeded: {
    (): Promise<void>;
  }
}

export class TransactionsView extends Component<PropTypes, {}> {

  componentDidMount() {
    this.props.fetchInitialTransactionsIfNeeded()
      .then(() => {
        console.log('done');
      })
      .catch(error => {
        console.error(error);
      })
  }

  render() {

    return (
      <span>Transactions</span>
    )
  }
}

export default connect(
  state => ({
    transactions: OrderedMap<number, Transaction>(),
  }),
  {
    fetchInitialTransactionsIfNeeded,
  }
)(TransactionsView)
