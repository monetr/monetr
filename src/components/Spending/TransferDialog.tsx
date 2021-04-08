import {
  Accordion,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormControl,
  Input,
  InputAdornment,
  InputLabel,
  List,
  ListItem,
  Typography
} from "@material-ui/core";
import SpendingSelectionList from 'components/Spending/SpendingSelectionList';
import Spending from 'data/Spending';
import React, { Component, Fragment } from "react";
import { connect } from 'react-redux';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';
import AccordionDetails from '@material-ui/core/AccordionDetails';
import AccordionSummary from '@material-ui/core/AccordionSummary';

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
}

interface State {
  from: Spending | null;
  to: Spending | null;
  selectionDialog: Target | null;
}

const SafeToSpend = new Spending({
  spendingId: -1, // Indicates that this is safe to spend.
  name: 'Safe-To-Spend',
});

enum Target {
  To,
  From,
}

class TransferDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    from: null,
    to: null,
    selectionDialog: null,
  };

  componentDidMount() {
    const { to, from } = this.props;
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

  openSelectionDialog = (target: Target) => () => {
    return this.setState({
      selectionDialog: target,
    });
  };

  renderSelectionDialog = () => {
    const { selectionDialog } = this.state;
    if (!selectionDialog) {
      return null;
    }

    let value: Spending | null = selectionDialog === Target.From ? this.state.from : this.state.to;

    if (value === SafeToSpend) {
      value = null;
    }

    let newValue: Spending | null = value;

    const onChange = (spending: Spending | null) => {
      newValue = spending;
    };

    const onOk = () => {
      console.log(newValue);
    };

    const onCancel = () => {
      return this.setState({
        selectionDialog: null,
      });
    };

    return (
      <Dialog open={ true } maxWidth="xs">
        <DialogTitle>
          Choose a goal or expense
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            Choose a goal or expense for your transfer.
          </DialogContentText>
          <SpendingSelectionList value={ newValue?.spendingId } onChange={ onChange }/>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={ onCancel }
          >
            Cancel
          </Button>
          <Button
            onClick={ onOk }
          >
            Ok
          </Button>
        </DialogActions>
      </Dialog>
    )
  };

  render() {
    const { isOpen, onClose } = this.props;
    return (
      <Fragment>
        { this.renderSelectionDialog() }
        <Dialog open={ isOpen } maxWidth="xs">
          <DialogTitle>
            Transfer Funds
          </DialogTitle>
          <DialogContent>
            <DialogContentText>
              Transfer funds to or from an expense or goal. This will allocate these funds to the destination so they
              can
              be put aside or used.
            </DialogContentText>
            <Accordion expanded={ false }>
              <AccordionSummary>
                <Typography
                  variant="h5"
                >
                  From
                </Typography>
              </AccordionSummary>
            </Accordion>
            <List>
              <ListItem
                key="from"
                button
                className="transfer-item"
                onClick={ this.openSelectionDialog(Target.From) }
              >
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
              </ListItem>
              <ListItem
                key="to"
                button
                className="transfer-item"
                onClick={ this.openSelectionDialog(Target.To) }
              >
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
              </ListItem>
            </List>
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
    };
  },
  {}
)(TransferDialog);
