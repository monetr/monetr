import {
  Accordion,
  AccordionDetails,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormControl,
  IconButton,
  Input,
  InputAdornment,
  InputLabel,
  Typography
} from "@material-ui/core";
import AccordionSummary from '@material-ui/core/AccordionSummary';
import { SwapVert } from '@material-ui/icons';
import classNames from 'classnames';
import SpendingSelectionList from 'components/Spending/SpendingSelectionList';
import Balance from 'data/Balance';
import Spending from 'data/Spending';
import React, { Component, Fragment } from "react";
import { connect } from 'react-redux';
import { getBalance } from 'shared/balances/selectors/getBalance';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

import './styles/TransferDialog.scss';

export interface PropTypes {
  initialFromSpendingId?: number;
  initialToSpendingId?: number;
  isOpen: boolean;
  onClose: { (): void }
}

interface WithConnectionPropTypes extends PropTypes {
  from: Spending | null;
  to: Spending | null;
  balance: Balance;
}

interface State {
  from: Spending | null;
  to: Spending | null;
  selectionDialog: Target | null;
}


enum Target {
  To,
  From,
}

let SafeToSpend = new Spending({
  spendingId: null, // Indicates that this is safe to spend.
  name: 'Safe-To-Spend',
});

class TransferDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    from: null,
    to: null,
    selectionDialog: null,
  };

  componentDidMount() {
    let { to, from, balance } = this.props;

    SafeToSpend.currentAmount = balance.safe;

    if (!to && from !== SafeToSpend) {
      to = SafeToSpend
    } else if (!from && to !== SafeToSpend) {
      from = SafeToSpend
    }

    this.setState({
      from,
      to,
    });
  }

  reverse = () => {
    this.setState(prevState => ({
      from: prevState.to,
      to: prevState.from,
    }));
  };

  doTransfer = () => {

  };

  renderSelection = (selection: Spending | null) => {
    if (!selection) {
      return (
        <div className="col-span-3 row-span-2">
          <Typography
            variant="h6"
          >
            Choose Goal or Expense
          </Typography>
        </div>
      )
    }

    return (
      <Fragment>
        <div className="col-span-3 row-span-1">
          <Typography
            variant="h6"
          >
            { selection.name }
          </Typography>
        </div>
        <div className="col-span-3 row-span-1 opacity-75">
          <Typography
            variant="body2"
          >
            { selection.getCurrentAmountString() } balance
          </Typography>
        </div>
      </Fragment>
    );
  };

  toggleExpanded = (target: Target) => () => {
    return this.setState(prevState => ({
      selectionDialog: prevState.selectionDialog === target ? null : target,
    }));
  }

  handleFromOnChange = (spending: Spending | null) => {
    return this.setState({
      from: spending ?? SafeToSpend,
    });
  };

  handleToOnChange = (spending: Spending | null) => {
    return this.setState({
      to: spending ?? SafeToSpend,
    });
  };

  render() {
    const { isOpen, onClose } = this.props;
    return (
      <Fragment>
        <Dialog open={ isOpen } maxWidth="xs">
          <DialogTitle>
            Transfer Funds
          </DialogTitle>
          <DialogContent className="p-5">
            <DialogContentText>
              Transfer funds to or from an expense or goal. This will allocate these funds to the destination so they
              can
              be put aside or used.
            </DialogContentText>
            <IconButton
              onClick={ this.reverse }
              color="primary"
              size="medium"
              className={ classNames('reverse-button transition-opacity', {
                'opacity-0': this.state.selectionDialog !== null,
                'opacity-100': this.state.selectionDialog === null,
              }) }
            >
              <SwapVert/>
            </IconButton>
            <div>
              <Accordion expanded={ this.state.selectionDialog === Target.From } className="transfer-item"
                         onChange={ this.toggleExpanded(Target.From) }>
                <AccordionSummary>
                  <div className='grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full'>
                    <div className="col-span-1 row-span-2">
                      <Typography
                        variant="h5"
                      >
                        From
                      </Typography>
                    </div>
                    { this.renderSelection(this.state.from) }
                  </div>
                </AccordionSummary>
                <AccordionDetails>
                  <SpendingSelectionList
                    value={ this.state.from?.spendingId }
                    onChange={ this.handleFromOnChange }
                    excludeIds={ this.state.to ? [this.state.to.spendingId] : null }
                    excludeSafeToSpend={ this.state.to === SafeToSpend }
                  />
                </AccordionDetails>
              </Accordion>
              <Accordion expanded={ this.state.selectionDialog === Target.To }
                         onChange={ this.toggleExpanded(Target.To) }>
                <AccordionSummary>
                  <div className='grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full'>
                    <div className="col-span-1 row-span-2">
                      <Typography
                        variant="h5"
                      >
                        To
                      </Typography>
                    </div>
                    { this.renderSelection(this.state.to) }
                  </div>
                </AccordionSummary>
                <AccordionDetails>
                  <SpendingSelectionList
                    value={ this.state.to?.spendingId }
                    onChange={ this.handleToOnChange }
                    excludeIds={ this.state.from ? [this.state.from.spendingId] : null }
                    excludeSafeToSpend={ this.state.from === SafeToSpend }
                  />
                </AccordionDetails>
              </Accordion>
            </div>
            <div className="w-full mt-5">
              <FormControl fullWidth>
                <InputLabel htmlFor="new-expense-amount">Amount</InputLabel>
                <Input
                  id="new-expense-amount"
                  name="amount"
                  value={ 0 }
                  onBlur={ () => {
                  } }
                  onChange={ () => {
                  } }
                  disabled={ false }
                  startAdornment={ <InputAdornment position="start">$</InputAdornment> }
                />
              </FormControl>
            </div>
          </DialogContent>
          <DialogActions>
            <Button
              onClick={ onClose }
            >
              Cancel
            </Button>
            <Button
              variant="outlined"
              color="primary"
              onClick={ this.doTransfer }
            >
              Transfer
            </Button>
          </DialogActions>
        </Dialog>
      </Fragment>
    );
  }
}

export default connect(
  (state, props: PropTypes) => {
    let from: Spending, to: Spending;

    switch (props.initialFromSpendingId) {
      case null:
      case undefined:
        break;
      case 0:
        from = SafeToSpend;
        break;
      default:
        from = getSpendingById(props.initialFromSpendingId)(state);
    }

    switch (props.initialToSpendingId) {
      case null:
      case undefined:
        break;
      case 0:
        to = SafeToSpend;
        break;
      default:
        to = getSpendingById(props.initialToSpendingId)(state);
    }

    return {
      from,
      to,
      balance: getBalance(state),
    };
  },
  {}
)(TransferDialog);
