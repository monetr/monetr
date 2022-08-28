import React, { Fragment } from 'react';
import {
  Button,
} from '@mui/material';
import { Formik, FormikErrors, FormikHelpers, FormikProps } from 'formik';
import moment from 'moment';
import { useSnackbar } from 'notistack';

import EditInProgressGoal from 'components/Goals/EditInProgressGoal';
import { useUpdateSpending } from 'hooks/spending';
import Spending from 'models/Spending';

interface editGoalForm {
  name: string;
  amount: number;
  dueDate: moment.Moment;
  fundingScheduleId: number;
}

interface Props {
  goal: Spending;
  hideView: () => void;
}

export default function EditGoalView(props: Props): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();

  const updateSpending = useUpdateSpending();

  function validateInput(_: editGoalForm): FormikErrors<editGoalForm> {
    return null;
  }

  async function submit(values: editGoalForm, { setSubmitting }: FormikHelpers<editGoalForm>): Promise<void> {
    setSubmitting(true);
    const updatedSpending = new Spending({
      ...props.goal,
      name: values.name,
      targetAmount: values.amount * 100,
      nextRecurrence: values.dueDate.startOf('day'),
      fundingScheduleId: values.fundingScheduleId,
    });

    return updateSpending(updatedSpending)
      .catch(error => void enqueueSnackbar(error?.response?.data?.error || 'Failed to update goal.', {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => setSubmitting(false));
  }

  const { goal, hideView } = props;

  function Contents({ formik }: { formik: FormikProps<editGoalForm> }): JSX.Element {
    if (goal.getGoalIsInProgress()) {
      return <EditInProgressGoal formik={ formik } hideView={ hideView } />;
    }

    // TODO Implement completed goals editing.
    return null;
  }

  const initial: editGoalForm = {
    name: goal.name,
    amount: goal.getTargetAmountDollars(),
    dueDate: goal.nextRecurrence,
    fundingScheduleId: goal.fundingScheduleId,
  };

  return (
    <Fragment>
      <Formik
        initialValues={ initial }
        validate={ validateInput }
        onSubmit={ submit }
      >
        { (formik: FormikProps<editGoalForm>) => (
          <form onSubmit={ formik.handleSubmit } className="h-full flex flex-col justify-between">
            <Contents formik={ formik } />
            <div>
              <Button
                className="w-full"
                variant="outlined"
                color="primary"
                disabled={ formik.isSubmitting }
                onClick={ formik.submitForm }
              >
                Update Goal
              </Button>
            </div>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
}
