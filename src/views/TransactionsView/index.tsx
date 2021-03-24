import { Box, Button, Card, Container, Grid, List } from "@material-ui/core";
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

export class TransactionsView extends Component<PropTypes, {}> {

  state = {
    selectedTransaction: 0,
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
    return this.setState({
      selectedTransaction: transactionId,
    })
  }

  renderTransaction = (transactionId: number) => {
    const { selectedTransaction } = this.state;
    return (
      <TransactionItem
        transactionId={ transactionId }
        selected={ transactionId === selectedTransaction }
        onClick={ this.selectTransaction }
      />
    );
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
                Test content
                Selected transaction { this.state.selectedTransaction }
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
