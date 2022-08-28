import React, { Fragment, useEffect, useState } from 'react';
import { SwapVert } from '@mui/icons-material';
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
  Typography,
} from '@mui/material';
import AccordionSummary from '@mui/material/AccordionSummary';
import classNames from 'classnames';
import { Formik, FormikHelpers } from 'formik';

import SpendingSelectionList from 'components/Spending/SpendingSelectionList';
import { useCurrentBalance } from 'hooks/balances';
import { useSpendingSink, useTransfer } from 'hooks/spending';
import Spending from 'models/Spending';

import './styles/TransferDialog.scss';

enum Target {
  To,
  From,
}

const SafeToSpend = new Spending({
  spendingId: null, // Indicates that this is safe to spend.
  name: 'Safe-To-Spend',
});

interface transferForm {
  amount: number;
}

const initialValues: transferForm = {
  amount: 0.00,
};

interface Props {
  initialFromSpendingId?: number;
  initialToSpendingId?: number;
  isOpen: boolean;
  onClose: { (): void }
}

export default function TransferDialog(props: Props): JSX.Element {
  if (props.initialFromSpendingId === 0 &&
      props.initialToSpendingId === 0) {
    throw new Error('the initial from and to spending IDs cannot both be 0');
  }
  const { result: spending } = useSpendingSink();
  const balance = useCurrentBalance();
  const transfer = useTransfer();

  interface ToFrom {
    to: Spending | null;
    from: Spending | null;
  }
  const [state, setState] = useState<ToFrom>({
    to: null,
    from: null,
  });

  const [selectionDialog, setSelectionDialog] = useState<Target | null>(null);

  const toggleExpanded = (target: Target) => () => setSelectionDialog(selectionDialog === target ? null : target);

  function handleToOnChange(spending: Spending | null) {
    setState({
      from: state.from,
      to: spending ?? SafeToSpend,
    });
  }

  function handleFromOnChange(spending: Spending | null) {
    setState({
      to: state.to,
      from: spending ?? SafeToSpend,
    });
  }

  useEffect(() => {
    let to: Spending, from: Spending;
    SafeToSpend.currentAmount = balance.safe;
    switch (props.initialFromSpendingId) {
      case null:
      case undefined:
      case 0: // a 0 for a spending ID represents safe-to-spend.
        from = SafeToSpend;
        break;
      default:
        from = spending.get(props.initialFromSpendingId);
    }

    switch (props.initialToSpendingId) {
      case null:
      case undefined:
      case 0: // a 0 for a spending ID represents safe-to-spend.
        to = SafeToSpend;
        break;
      default:
        to = spending.get(props.initialToSpendingId);
    }

    // Then persist these spending objects to state to be used.
    setState({
      to,
      from,
    });
  }, [props.initialToSpendingId, props.initialFromSpendingId]);

  function reverse(): void {
    setState({
      from: state.to,
      to: state.from,
    });
  }

  function doTransfer(values: transferForm, { setSubmitting, setFieldError }: FormikHelpers<transferForm>) {
    const { to, from } = state;

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

    return transfer(fromId, toId, amount)
      .then(() => onClose());
  }

  function SpendingSelection(props: { selection: Spending | null }): JSX.Element {
    const { selection } = props;
    if (!selection) {
      return (
        <div className="col-span-3 row-span-2">
          <Typography
            variant="h6"
          >
            Choose Goal or Expense
          </Typography>
        </div>
      );
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
  }

  const { isOpen, onClose } = props;
  return (
    <Formik
      initialValues={ initialValues }
      onSubmit={ doTransfer }
    >
      { ({
        values,
        handleChange,
        handleBlur,
        handleSubmit,
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
                onClick={ reverse }
                color="primary"
                size="medium"
                className={ classNames('reverse-button transition-opacity', {
                  'opacity-0': selectionDialog !== null,
                  'opacity-100': selectionDialog === null,
                }) }
              >
                <SwapVert />
              </IconButton>
              <div>
                <Accordion expanded={ selectionDialog === Target.From } className="transfer-item"
                  onChange={ toggleExpanded(Target.From) }>
                  <AccordionSummary>
                    <div className='grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full'>
                      <div className="col-span-1 row-span-2">
                        <Typography
                          variant="h5"
                        >
                          From
                        </Typography>
                      </div>
                      <SpendingSelection selection={ state.from } />
                    </div>
                  </AccordionSummary>
                  <AccordionDetails>
                    <SpendingSelectionList
                      disabled={ isSubmitting }
                      value={ state.from?.spendingId }
                      onChange={ handleFromOnChange }
                      excludeIds={ state.to ? [state.to.spendingId] : null }
                      excludeSafeToSpend={ state.to === SafeToSpend }
                    />
                  </AccordionDetails>
                </Accordion>
                <Accordion expanded={ selectionDialog === Target.To }
                  onChange={ toggleExpanded(Target.To) }>
                  <AccordionSummary>
                    <div className='grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full'>
                      <div className="col-span-1 row-span-2">
                        <Typography
                          variant="h5"
                        >
                          To
                        </Typography>
                      </div>
                      <SpendingSelection selection={ state.to } />
                    </div>
                  </AccordionSummary>
                  <AccordionDetails>
                    <SpendingSelectionList
                      disabled={ isSubmitting }
                      value={ state.to?.spendingId }
                      onChange={ handleToOnChange }
                      excludeIds={ state.from ? [state.from.spendingId] : null }
                      excludeSafeToSpend={ state.from === SafeToSpend }
                    />
                  </AccordionDetails>
                </Accordion>
              </div>
              <div className="w-full mt-5">
                <FormControl fullWidth>
                  <InputLabel htmlFor="new-expense-amount">Amount</InputLabel>
                  <Input
                    autoFocus={ true }
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

