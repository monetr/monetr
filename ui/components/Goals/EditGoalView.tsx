import { connect } from "react-redux";
import React, { Component, Fragment } from "react";
import Spending from "models/Spending";
import { Formik, FormikErrors, FormikHelpers, FormikProps } from "formik";
import { getSelectedGoal } from "shared/spending/selectors/getSelectedGoal";
import {
  Button,
  Divider,
  FormControl,
  IconButton,
  Input,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  Typography
} from "@material-ui/core";
import { ArrowBack, DeleteOutline } from "@material-ui/icons";
import moment from "moment";
import updateSpending from "shared/spending/actions/updateSpending";
import { KeyboardDatePicker, MuiPickersUtilsProvider } from "@material-ui/pickers";
import MomentUtils from "@date-io/moment";
import { getFundingSchedules } from "shared/fundingSchedules/selectors/getFundingSchedules";
import FundingSchedule from "models/FundingSchedule";
import { Map } from 'immutable';

export interface PropTypes {
  hideView: { (): void }
}

interface WithConnectionPropTypes extends PropTypes {
  goal: Spending;
  fundingSchedules: Map<number, FundingSchedule>;
  updateSpending: { (spending: Spending): Promise<void> }
}

interface editGoalForm {
  name: string;
  amount: number;
  dueDate: moment.Moment;
  fundingScheduleId: number;
}

export class EditGoalView extends Component<WithConnectionPropTypes, any> {

  validateInput = (values: editGoalForm): FormikErrors<any> => {
    return null;
  };

  submit = (values: editGoalForm, { setSubmitting }: FormikHelpers<editGoalForm>) => {
    const { goal, updateSpending } = this.props;

    const updatedSpending = new Spending({
      ...goal,
      name: values.name,
      targetAmount: values.amount * 100,
      nextRecurrence: values.dueDate.startOf('day'),
      fundingScheduleId: values.fundingScheduleId,
    });

    return updateSpending(updatedSpending)
      .then(() => {
        setSubmitting(false);
      })
      .catch(error => {
        setSubmitting(false);

        this.setState({
          error: error.response.data.error,
        });
      });
  };

  renderContents = (formik: FormikProps<editGoalForm>) => {
    const { goal } = this.props;

    if (goal.getGoalIsInProgress()) {
      return this.renderInProgress(formik);
    }

    return this.renderComplete();
  };

  renderTopBar = () => {
    return (
      <Fragment>
        <div className="w-full h-12">
          <div className="grid grid-cols-6 grid-rows-1 grid-flow-col">
            <div className="col-span-1">
              <IconButton
                onClick={ () => this.props.hideView() }
              >
                <ArrowBack/>
              </IconButton>
            </div>
            <div className="col-span-4 flex justify-center items-center">
              <Typography
                variant="h6"
              >
                Edit Goal
              </Typography>
            </div>
            <div className="col-span-1">
              <IconButton disabled>
                <DeleteOutline/>
              </IconButton>
            </div>
          </div>
        </div>
        <Divider/>
      </Fragment>
    )
  };

  renderInProgress = (formik: FormikProps<editGoalForm>) => {
    const { fundingSchedules } = this.props;

    return (
      <div className="w-full h-full flex-grow">
        { this.renderTopBar() }

        <FormControl fullWidth className="mt-5">
          <InputLabel htmlFor="edit-goal-name">Goal Name</InputLabel>
          <Input
            autoFocus={ true }
            id="edit-goal-name"
            name="name"
            value={ formik.values.name }
            onBlur={ formik.handleBlur }
            onChange={ formik.handleChange }
            disabled={ formik.isSubmitting }
          />
        </FormControl>
        <FormControl fullWidth className="mt-5">
          <InputLabel htmlFor="edit-goal-amount">Target Amount</InputLabel>
          <Input
            id="edit-goal-amount"
            name="amount"
            value={ formik.values.amount }
            onBlur={ formik.handleBlur }
            onChange={ formik.handleChange }
            disabled={ formik.isSubmitting }
            startAdornment={ <InputAdornment position="start">$</InputAdornment> }
          />
        </FormControl>
        <KeyboardDatePicker
          className="mt-5"
          fullWidth
          minDate={ moment().subtract('1 day') }
          name="date"
          margin="normal"
          id="date-picker-dialog"
          label="When do you need it by?"
          format="MM/DD/yyyy"
          value={ formik.values.dueDate }
          onChange={ (value) => formik.setFieldValue('dueDate', value) }
          disabled={ formik.isSubmitting }
          KeyboardButtonProps={ {
            'aria-label': 'change date',
          } }
        />
        <FormControl fullWidth className="mt-5">
          <InputLabel id="edit-funding-schedule-label">Funding Schedule</InputLabel>
          <Select
            labelId="edit-funding-schedule-label"
            id="edit-funding-schedule"
            name="fundingScheduleId"
            value={ formik.values.fundingScheduleId }
            onChange={ formik.handleChange }
            disabled={ formik.isSubmitting }
          >
            { fundingSchedules.map(item => (
              <MenuItem
                key={ item.fundingScheduleId }
                value={ item.fundingScheduleId }
              >
                { item.name }
              </MenuItem>
            )).toArray() }
          </Select>
        </FormControl>
      </div>
    )
  };

  renderComplete = () => {
    return null;
  };

  render() {
    const { goal } = this.props;
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
          validate={ this.validateInput }
          onSubmit={ this.submit }
        >
          { (formik: FormikProps<editGoalForm>) => (
            <form onSubmit={ formik.handleSubmit } className="h-full flex flex-col justify-between">
              <MuiPickersUtilsProvider utils={ MomentUtils }>
                { this.renderContents(formik) }
              </MuiPickersUtilsProvider>
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
    )
  }
}

export default connect(
  (state) => ({
    goal: getSelectedGoal(state),
    fundingSchedules: getFundingSchedules(state),
  }),
  {
    updateSpending,
  }
)(EditGoalView);
