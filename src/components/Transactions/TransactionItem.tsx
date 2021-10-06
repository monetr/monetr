import { Button, Chip, Divider, ListItem, Paper, Popover, Typography } from "@material-ui/core";
import classnames from 'classnames';
import Spending from "data/Spending";
import Transaction from "data/Transaction";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import { getSpendingById } from "shared/spending/selectors/getSpendingById";
import selectTransaction from "shared/transactions/actions/selectTransaction";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";

import './styles/TransactionItem.scss';
import { getTransactionIsSelected } from "shared/transactions/selectors/getTransactionIsSelected";
import SelectButton from "components/SelectyBoi/SelectButton";
import SpendingSelectionList from "components/Spending/SpendingSelectionList";
import updateTransaction from "shared/transactions/actions/updateTransaction";

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  spending?: Spending;
  isSelected: boolean;
  selectTransaction: { (transactionId: number): void }
  updateTransaction: (transaction: Transaction) => Promise<void>;
}

interface State {
  spentFromAnchorEl: Element | null;
  spentFromWidth: number | null;
  nameAnchorEl: Element | null;
  nameWidth: number | null;
}

export class TransactionItem extends Component<WithConnectionPropTypes, State> {

  state = {
    spentFromAnchorEl: null,
    spentFromWidth: 0,
    nameAnchorEl: null,
    nameWidth: 0,
  };

  getSpentFromString() {
    const { spending, transaction, updateTransaction } = this.props;
    const { spentFromAnchorEl } = this.state;

    if (transaction.getIsAddition()) {
      return null;
    }

    const updateSpentFrom = (selection: Spending | null) => {
      const spendingId = selection ? selection.spendingId : null;

      if (spendingId === transaction.spendingId) {
        return Promise.resolve();
      }

      const updatedTransaction = new Transaction({
        ...transaction,
        spendingId: spendingId,
      });

      return updateTransaction(updatedTransaction)
        .catch(error => alert(error));
    };

    const openPopover = (event: { currentTarget: Element }) => {
      this.setState({
        spentFromAnchorEl: event.currentTarget,
        spentFromWidth: event.currentTarget.clientWidth,
        nameAnchorEl: null,
        nameWidth: null,
      });
    };

    const closePopover = () => this.setState({
      spentFromAnchorEl: null,
      spentFromWidth: null,
      nameAnchorEl: null,
      nameWidth: null,
    });

    return (
      <Fragment>
        <SelectButton
          open={ Boolean(spentFromAnchorEl) }
          onClick={ openPopover }
        >
          <span className="mr-1 opacity-50">
            Spent From
          </span>
          <span className={ classnames('overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap', {
            'opacity-50': !spending,
          }) }>
            { spending ? spending.name : 'Safe-To-Spend' }
          </span>
        </SelectButton>
        <Popover
          id={ `transaction-spent-from-popover-${ transaction.transactionId }` }
          open={ Boolean(spentFromAnchorEl) }
          anchorEl={ spentFromAnchorEl }
          onClose={ closePopover }
          anchorOrigin={ {
            vertical: 'bottom',
            horizontal: 'left',
          } }
          transformOrigin={ {
            vertical: 'top',
            horizontal: 'left',
          } }
        >
          <Paper
            style={ { width: `${ this.state.spentFromWidth }px` } } 
            className="p-0 overflow-auto min-w-96 max-h-96"
          >
            <SpendingSelectionList 
              value={ transaction.spendingId } 
              onChange={ updateSpentFrom }
            />
          </Paper>
        </Popover>
      </Fragment>
    )
  }

  renderTransactionName = () => {
    const { transaction } = this.props;
    const { nameAnchorEl, nameWidth } = this.state;

    const openPopover = (event: { currentTarget: Element }) => {
      this.setState({
        spentFromAnchorEl: null,
        spentFromWidth: null,
        nameAnchorEl: event.currentTarget,
        nameWidth: event.currentTarget.clientWidth,
      });
    };

    const closePopover = () => this.setState({
      spentFromAnchorEl: null,
      spentFromWidth: null,
      nameAnchorEl: null,
      nameWidth: null,
    });

    return (
      <Fragment>
        <SelectButton
          open={ Boolean(nameAnchorEl) }
          onClick={ openPopover }
        >
          <span className={ classnames('overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap') }>
            { transaction.getTitle() }
          </span>
        </SelectButton>
        <Popover
          id={ `transaction-name-popover-${ transaction.transactionId }` }
          open={ Boolean(nameAnchorEl) }
          anchorEl={ nameAnchorEl }
          onClose={ closePopover }
          anchorOrigin={ {
            vertical: 'bottom',
            horizontal: 'left',
          } }
          transformOrigin={ {
            vertical: 'top',
            horizontal: 'left',
          } }
        >
          <Paper style={ { width: `${ nameWidth }px` } } className="p-0 p-2 overflow-auto min-w-96 max-h-96">
            <div className="flex flex-row pt-1 pb-1">
              <div className="flex-auto">
                <p className="w-full text-lg font-semibold">Name:</p>
                <Divider />
                <span className="text-lg">{ transaction.getTitle() }</span>
              </div>
              <div className="flex-none h-full">
                <Button>Test</Button>
              </div>
            </div>
            <div className="pt-1 pb-1">
              <p className="w-full text-lg font-semibold">Original Transaction Name:</p>
              <Divider />
              <span className="text-lg">{ transaction.originalName }</span>
            </div>
          </Paper>
        </Popover>
      </Fragment>
    )
  };

  handleClick = () => {
    return this.props.selectTransaction(this.props.transactionId);
  }

  render() {
    const { transaction, isSelected } = this.props;

    return (
      <Fragment>
        <ListItem className={ classnames('transactions-item h-12', {
          'selected': false,
        }) } role="transaction-row">
          <div className="flex flex-row w-full">
            <p
              className="flex-shrink w-2/5 pr-1 overflow-hidden font-semibold transaction-item-name overflow-ellipsis flex-nowrap whitespace-nowrap place-self-center"
            >
              { this.renderTransactionName() }
            </p>

            <p
              className="flex-auto pr-1 overflow-hidden transaction-expense-name overflow-ellipsis flex-nowrap whitespace-nowrap"
            >
              { this.getSpentFromString() }
            </p>
            <div className="flex items-center flex-none w-1/5">
              { transaction.isPending && <Chip label="Pending" className="self-center align-middle"/> }
              <div className="flex justify-end w-full">
                <Typography className={ classnames('amount align-middle self-center place-self-center', {
                  'addition': transaction.getIsAddition(),
                }) }>
                  <b>{ transaction.getAmountString() }</b>
                </Typography>
              </div>
            </div>
          </div>
        </ListItem>
        <Divider/>
      </Fragment>
    )
  }
}

export default connect(
  (state, props: PropTypes) => {
    const transaction = getTransactionById(props.transactionId)(state);
    const isSelected = getTransactionIsSelected(props.transactionId)(state);

    return {
      transaction,
      isSelected,
      spending: getSpendingById(transaction.spendingId)(state),
    }
  },
  {
    selectTransaction,
    updateTransaction,
  }
)(TransactionItem)
