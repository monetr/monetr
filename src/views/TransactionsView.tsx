import { OrderedMap } from "immutable";
import Transaction from "data/Transaction";
import React, { Component } from "react";
import { connect } from "react-redux";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import { getTransactions } from "shared/transactions/selectors/getTransactions";
import { Box, Card, List, ListItem, ListItemText, Typography } from "@material-ui/core";

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

  renderTransactions = () => {
    const { transactions } = this.props;

    return transactions.toArray().map(([_, item]) => this.renderTransaction(item));
  }

  renderTransaction = (transaction: Transaction) => {
    return (
      <ListItem
        dense
        button
        key={ transaction.transactionId.toString() }
        alignItems="flex-start"
      >
        <ListItemText
          primary={ transaction.name }
          secondary={
            <React.Fragment>
              <Typography
                component="span"
                variant="body2"
                color="textPrimary"
              >
                { transaction.getAmountString() }
              </Typography>
            </React.Fragment>
          }
        >

        </ListItemText>
      </ListItem>
    )
  };

  render() {

    return (
      <Box m={ 6 }>
        <Card elevation={ 6 }>
          <List>
            { this.renderTransactions() }
          </List>
        </Card>
      </Box>
    )
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
