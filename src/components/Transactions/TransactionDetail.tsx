import { Button, Chip, Divider, Typography } from "@material-ui/core";
import Spending from "data/Spending";
import Transaction from "data/Transaction";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import { getSpending } from "shared/spending/selectors/getSpending";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";
import { Map } from 'immutable';
import classnames from 'classnames';

import './styles/TransactionDetail.scss';

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  spending: Map<number, Spending>;
}

class TransactionDetailView extends Component<WithConnectionPropTypes, {}> {

  render() {
    const { transaction } = this.props;

    return (
      <div className="w-full p-5 transaction-detail">
        <div className="grid grid-cols-1 grid-rows-2 grid-flow-col gap-1 w-auto">
          <Typography variant="h5">
            { transaction.date.format('MMMM Do, YYYY') }
          </Typography>
          <Typography variant="h6" className={ classnames('amount', {
            'addition': transaction.getIsAddition(),
          }) }>
            { transaction.getAmountString() }
          </Typography>
        </div>
        <Divider className="mt-5 mb-5" />

        <div className="grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full">
          <div className="col-span-3 row-span-1">
            <Typography variant="h5">Name</Typography>
          </div>
          <div className="col-span-3 row-span-1">
            <Typography>{ transaction.name }</Typography>
          </div>
          <div className="col-span-1 row-span-2 justify-end flex">
            <Button color="primary" className="align-middle self-center">Change</Button>
          </div>
        </div>
        <Divider className="mt-5 mb-5" />

        <div className="grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full">
          <div className="col-span-3 row-span-1">
            <Typography variant="h5">Categories</Typography>
          </div>
          <div className="col-span-3 row-span-1">
            {
              transaction.categories.map(cat => (
                <Chip
                  className="mr-1 mb-1"
                  key={ cat }
                  label={ cat }
                  variant="outlined"
                />
              ))
            }
          </div>
          <div className="col-span-1 row-span-2 justify-end flex">
            <Button color="primary" className="align-middle self-center">Change</Button>
          </div>
        </div>
        <Divider className="mt-5 mb-5" />

        {
          // Deposits are not spent from anything, so we don't want to show this for deposits.
          !transaction.getIsAddition() &&
          <Fragment>
            <div className="grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full">
              <div className="col-span-3 row-span-1">
                <Typography variant="h5">Spent From</Typography>
              </div>
              <div className="col-span-3 row-span-1">
                <Typography>Safe-To-Spend</Typography>
              </div>
              <div className="col-span-1 row-span-2 justify-end flex">
                <Button color="primary" className="align-middle self-center">Change</Button>
              </div>
            </div>
            <Divider className="mt-5 mb-5" />
          </Fragment>
        }
      </div>
    );
  }
}

export default connect(
  (state, props: PropTypes) => {
    const transaction = getTransactionById(props.transactionId)(state);

    return {
      transaction,
      spending: getSpending(state),
    };
  },
  {}
)(TransactionDetailView);
