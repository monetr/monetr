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
import transfer from 'shared/spending/actions/transfer';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';
import { Formik, FormikErrors, FormikHandlers, FormikHelpers } from "formik";

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
  transfer: { (from: number | null, to: number | null, amount: number): Promise<void> }
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

interface transferForm {
  amount: number;
}

const initialValues: transferForm = {
  amount: 0.00,
};

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

  doTransfer = (values: transferForm, { setSubmitting, setFieldError }: FormikHelpers<transferForm>) => {
    const { to, from } = this.state;

    if (values.amount <= 0) {
      setFieldError('amount', 'Amount must be greater than 0');
      setSubmitting(false);
      return null;
    }

    if (values.amount > from.currentAmount) {
      setFieldError('amount', 'Amount must be less than or equal to the source amount');
      setSubmitting(false);
      return null;
    }

    const fromId = from === SafeToSpend ? null : from.spendingId;
    const toId = to === SafeToSpend ? null : to.spendingId;
    const amount = Math.round(values.amount * 100);

    return this.props.transfer(fromId, toId, amount)
      .then(() => {
        this.props.onClose();
      })
      .catch(error => {
        console.error(error);
        setSubmitting(false);
      });
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
      <Formik
        initialValues={ initialValues }
        onSubmit={ this.doTransfer }
      >
        { ({
          values,
          errors,
          touched,
          handleChange,
          handleBlur,
          handleSubmit,
          setFieldValue,
          isSubmitting,
          submitForm,
        }) => (
          <form onSubmit={ handleSubmit }>
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
                        disabled={ isSubmitting }
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
                        disabled={ isSubmitting }
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
                      type="number"
                      id="new-expense-amount"
                      name="amount"
                      value={ values.amount }
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      disabled={ isSubmitting }
                      startAdornment={ <InputAdornment position="start">$</InputAdornment> }
                    />
                  </FormControl>
                </div>
              </DialogContent>
              <DialogActions>
                <Button
                  onClick={ onClose }
                  disabled={ isSubmitting }
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="outlined"
                  color="primary"
                  disabled={ isSubmitting }
                  onClick={ submitForm }
                >
                  Transfer
                </Button>
              </DialogActions>
            </Dialog>
          </form>
        )}
      </Formik>
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
  {
    transfer,
  }
)(TransferDialog);
