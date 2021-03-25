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

  selectTransaction = (transactionId: number) => {
    return this.setState(prevState => ({
      // This logic will make it so that if the selectTransaction method is called again for a transaction that is
      // already selected, then the selection will be toggled.
      selectedTransaction: transactionId === prevState.selectedTransaction ? 0 : transactionId
    }));
  }

  renderTransaction = (transactionId: number) => {
    const { selectedTransaction } = this.state;
    return (
      <TransactionItem
        key={ transactionId }
        transactionId={ transactionId }
        selected={ transactionId === selectedTransaction }
        onClick={ this.selectTransaction }
      />
    );
  };

  renderTransactionDetailView = () => {
    const { selectedTransaction } = this.state;

    if (selectedTransaction) {
      return (
        <TransactionDetailView transactionId={ selectedTransaction } />
      );
    }


    return (
      <div className="flex justify-center place-content-center">
        <Typography className="pt-10">Nothing here...</Typography>
      </div>
    )
  };

  render() {
    return (
      <div className="minus-nav">
        <div className="flex flex-col h-full p-10 max-h-full overflow-y-scroll">
          <div className="grid grid-cols-3 gap-4 flex-grow">
            <div className="col-span-2">
              <Card elevation={ 4 } className="w-full overflow-scroll table">
                <List disablePadding className="w-full">
                  { this.renderTransactions() }
                </List>
              </Card>
            </div>
            <div className="">
              <Card elevation={ 4 } className="h-full w-full">
                { this.renderTransactionDetailView() }
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
