import React, { Fragment } from 'react';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material';
import { Formik, FormikErrors, FormikHelpers } from 'formik';

import FundingScheduleSelectionList from 'components/FundingSchedules/FundingScheduleSelectionList';
import { useUpdateSpending } from 'hooks/spending';
import Spending from 'models/Spending';

interface Props {
  spending: Spending;
  onClose: () => void;
  isOpen: boolean;
}

interface editFundingScheduleForm {
  fundingScheduleId: number;
}

export default function EditFundingScheduleDialog(props: Props): JSX.Element {
  const updateSpending = useUpdateSpending();


  function validateInput(_: editFundingScheduleForm): FormikErrors<editFundingScheduleForm> {
    return null;
  }

  async function submit(
    values: editFundingScheduleForm,
    { setSubmitting }: FormikHelpers<editFundingScheduleForm>,
  ): Promise<void> {
    setSubmitting(true);
    const updatedSpending = new Spending({
      ...props.spending,
      fundingScheduleId: values.fundingScheduleId,
    });

    return updateSpending(updatedSpending)
      .then(() => props.onClose())
      .finally(() => setSubmitting(false));
  }

  const initial: editFundingScheduleForm = {
    fundingScheduleId: props.spending.fundingScheduleId,
  };

  return (
    <Fragment>
      <Formik
        initialValues={ initial }
        validate={ validateInput }
        onSubmit={ submit }
      >
        { ({
          values,
          handleSubmit,
          setFieldValue,
          isSubmitting,
          submitForm,
        }) => (
          <form onSubmit={ handleSubmit }>
            <Dialog open={ props.isOpen } maxWidth="sm">
              <DialogTitle>
                Edit expense funding schedule
              </DialogTitle>
              <DialogContent>
                <DialogContentText>
                  Change the funding schedule you want to use to make contributions to this expense.
                  <br/>
                  <span className="text-sm">
                    Note: The amount contributed to the expense cannot be changed here, it is calculated based on how much
                    is needed vs how frequently it gets funded.
                  </span>
                </DialogContentText>
                <div className="mt-2">
                  <FundingScheduleSelectionList
                    initialValue={ values.fundingScheduleId }
                    disabled={ isSubmitting }
                    onChange={ value => setFieldValue('fundingScheduleId', value.fundingScheduleId) }
                  />
                </div>
              </DialogContent>
              <DialogActions>
                <Button
                  color="secondary"
                  onClick={ props.onClose }
                >
                  Cancel
                </Button>
                <Button
                  disabled={ isSubmitting }
                  onClick={ submitForm }
                  color="primary"
                  type="submit"
                >
                  Update
                </Button>
              </DialogActions>
            </Dialog>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
}
