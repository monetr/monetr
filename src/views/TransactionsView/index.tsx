import { Card, Divider, List, ListSubheader, Typography } from "@material-ui/core";
import TransactionDetailView from "components/Transactions/TransactionDetail";
import TransactionItem from "components/Transactions/TransactionItem";
import React, { Component, Fragment } from "react";
import { connect } from "react-redux";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import Transaction from "data/Transaction";
import { getTransactions } from "shared/transactions/selectors/getTransactions";
import { Map } from 'immutable';

import './styles/TransactionsView.scss';

interface PropTypes {
  transactions: Map<number, Transaction>;
  fetchInitialTransactionsIfNeeded: {
    (): Promise<void>;
  }
}

export class TransactionsView extends Component<PropTypes, any> {

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
    const { transactions } = this.props;

    return transactions
      .groupBy(transaction => transaction.date.format('MMMM Do'))
      .map((transactions, group) => (
        <li key={ group }>
          <ul>
            <Fragment>
              <ListSubheader className="bg-white pl-0 pr-0">
                <Typography className="ml-2 font-semibold opacity-75 text-base">{ group }</Typography>
                <Divider/>
              </ListSubheader>
            </Fragment>
            { transactions.map(transaction => (
              <TransactionItem key={ transaction.transactionId }
                               transactionId={ transaction.transactionId }
              />)).valueSeq().toArray() }
          </ul>
        </li>
      ))
      .valueSeq()
      .toArray();
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
                <TransactionDetailView/>
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
    transactions: getTransactions(state),
  }),
  {
    fetchInitialTransactionsIfNeeded,
  }
)(TransactionsView)
