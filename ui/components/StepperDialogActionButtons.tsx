import React, { Fragment } from 'react';
import { Button } from '@mui/material';

export enum StepperStep {
  First,
  Other,
  Last,
}

interface Props {
  isSubmitting: boolean;
  submitForm: () => Promise<unknown>;
  onClose: () => void;
  previousStep: () => void;
  nextStep: () => void;
  canNextStep: boolean;
  step: StepperStep;
}

function StepperDialogActionButtons(props: Props): JSX.Element {
  const { isSubmitting, submitForm, onClose, previousStep, nextStep, canNextStep, step } = props;
  const CancelButton = () => (
    <Button color="secondary" onClick={ onClose }>
      Cancel
    </Button>
  );

  const PreviousButton = () => (
    <Button
      disabled={ isSubmitting }
      color="secondary"
      onClick={ previousStep }
    >
      Previous
    </Button>
  );

  const NextButton = () => (
    <Button
      color="primary"
      onClick={ nextStep }
      disabled={ !canNextStep }
    >
      Next
    </Button>
  );

  const SubmitButton = () => (
    <Button
      disabled={ isSubmitting }
      onClick={ submitForm }
      color="primary"
      type="submit"
    >
      Create
    </Button>
  );

  switch (step) {
    case StepperStep.First:
      return (
        <Fragment>
          <CancelButton />
          <NextButton />
        </Fragment>
      );
    case StepperStep.Last:
      return (
        <Fragment>
          <PreviousButton />
          <SubmitButton />
        </Fragment>
      );
    default:
      return (
        <Fragment>
          <PreviousButton />
          <NextButton />
        </Fragment>
      );
  }
}

export default React.memo(StepperDialogActionButtons);
