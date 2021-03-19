import { Box, Button, Card, Container, Grid } from "@material-ui/core";
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
      <TransactionItem transactionId={ transactionId }/>
    );
  };

  render() {

    return (
      <Box m={ 6 }>
        <Container maxWidth="lg">
          <Grid container spacing={ 2 } justify={ "flex-start" }>
            <Grid item md={ 8 }>
              <Card elevation={ 6 }>
                <div className="transactions-view">
                  { this.renderTransactions() }
                </div>
              </Card>
            </Grid>
            <Grid item sm>
              <Card elevation={ 6 }>
                <Button>Test</Button>
              </Card>
            </Grid>
          </Grid>
        </Container>
      </Box>
    )
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
