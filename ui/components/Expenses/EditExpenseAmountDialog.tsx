import React, { Fragment, useState } from 'react';
import {
  Alert,
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
  Snackbar,
} from '@mui/material';
import { Formik, FormikErrors, FormikHelpers } from 'formik';

import { useUpdateSpending } from 'hooks/spending';
import Spending from 'models/Spending';

interface editSpendingAmountForm {
  amount: number;
}

export interface Props {
  spending: Spending;
  onClose: { (): void };
  isOpen: boolean;
}

export default function EditExpenseAmountDialog(props: Props): JSX.Element {
  const [error, setError] = useState<string|null>(null);
  const updateSpending = useUpdateSpending();

  const initialValues: editSpendingAmountForm = {
    amount: props.spending.getTargetAmountDollars(),
  };

  function validateInput(_: editSpendingAmountForm): FormikErrors<editSpendingAmountForm> {
    return null;
  }

  async function submit(
    values: editSpendingAmountForm,
    { setSubmitting }: FormikHelpers<editSpendingAmountForm>,
  ): Promise<void> {
    setSubmitting(true);
    const updatedSpending = new Spending({
      ...props.spending,
      targetAmount: values.amount * 100,
    });

    return updateSpending(updatedSpending)
      .then(() => props.onClose())
      .catch(error => setError(error.response.data.error))
      .finally(() => setSubmitting(false));
  }

  function ErrorMaybe(): JSX.Element {
    if (!error) {
      return null;
    }

    const onClose = () => setError(null);

    return (
      <Snackbar open autoHideDuration={ 6000 } onClose={ onClose }>
        <Alert onClose={ onClose } severity="error">
          { error }
        </Alert>
      </Snackbar>
    );
  }

  return (
    <Fragment>
      <ErrorMaybe />
      <Formik
        initialValues={ initialValues }
        validate={ validateInput }
        onSubmit={ submit }
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
            <Dialog open={ props.isOpen } maxWidth='xs'>
              <DialogTitle>
                Edit amount
              </DialogTitle>
              <DialogContent>
                <DialogContentText>
                  Edit the amount you want to put aside for this goal or expense. This will change how much is put
                  aside
                  each time you are paid as well to make sure you have enough by your target date.
                </DialogContentText>
                <FormControl fullWidth>
                  <InputLabel htmlFor="edit-expense-amount">Target Amount</InputLabel>
                  <Input
                    autoFocus={ true }
                    id="edit-expense-amount"
                    name="amount"
                    value={ values.amount }
                    onBlur={ handleBlur }
                    onChange={ handleChange }
                    disabled={ isSubmitting }
                    startAdornment={ <InputAdornment position="start">$</InputAdornment> }
                  />
                </FormControl>
              </DialogContent>
              <DialogActions>
                <Button
                  disabled={ isSubmitting }
                  onClick={ props.onClose }
                >
                  Cancel
                </Button>
                <Button
                  disabled={ isSubmitting }
                  color="primary"
                  onClick={ submitForm }
                >
                  Save
                </Button>
              </DialogActions>
            </Dialog>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
}
