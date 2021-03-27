import { Box, Button, Card, Container, Grid, List, Typography } from "@material-ui/core";
import TransactionDetailView from "components/Transactions/TransactionDetail";
import TransactionItem from "components/Transactions/TransactionItem";
import React, { Component } from "react";
import { connect } from "react-redux";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import { getTransactionIds } from "shared/transactions/selectors/getTransactionIds";

import './styles/TransactionsView.scss';

interface PropTypes {
  transactionIds: number[];
  fetchInitialTransactionsIfNeeded: {
    (): Promise<void>;
  }
}

interface State {
  selectedTransaction: number;
}

export class TransactionsView extends Component<PropTypes, State> {

  state = {
    selectedTransaction: 0
  };

  componentDidMount() {
    this.props.fetchInitialTransactionsIfNeeded()
      .then(() => {
        console.log('done');
      })
      .catch(error => {
        console.error(error);
      })
  }

  renderTransactions = () => {
    const { transactionIds } = this.props;

    return transactionIds.map(transactionId => this.renderTransaction(transactionId));
  }

  renderTransaction = (transactionId: number) => {
    return (
      <TransactionItem
        key={ transactionId }
        transactionId={ transactionId }
      />
    );
  };

  render() {
    return (
      <div className="minus-nav">
        <div className="flex flex-col h-full p-10 max-h-full">
          <div className="grid grid-cols-3 gap-4 flex-grow">
            <div className="col-span-2">
              <Card elevation={ 4 } className="w-full transaction-list">
                <List disablePadding className="w-full">
                  { this.renderTransactions() }
                </List>
              </Card>
            </div>
            <div className="">
              <Card elevation={ 4 } className="h-full w-full">
                <TransactionDetailView />
              </Card>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default connect(
  state => ({
    transactionIds: getTransactionIds(state),
  }),
  {
    fetchInitialTransactionsIfNeeded,
  }
)(TransactionsView)
