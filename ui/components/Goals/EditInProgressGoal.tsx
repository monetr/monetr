import React from 'react';
import { ArrowBack, DeleteOutline } from '@mui/icons-material';
import { DatePicker } from '@mui/lab';
import {
  Divider,
  FormControl,
  IconButton,
  Input,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  Typography,
} from '@mui/material';
import { FormikProps } from 'formik';
import moment from 'moment';

import { useFundingSchedules } from 'hooks/fundingSchedules';

interface editGoalForm {
  name: string;
  amount: number;
  dueDate: moment.Moment;
  fundingScheduleId: number;
}

interface Props {
  formik: FormikProps<editGoalForm>
  hideView: () => void;
}

export default function EditInProgressGoal(props: Props): JSX.Element {
  const { formik } = props;
  const fundingSchedules = useFundingSchedules();
  return (
    <div className="w-full h-full flex-grow">
      <div className="w-full h-12">
        <div className="grid grid-cols-6 grid-rows-1 grid-flow-col">
          <div className="col-span-1">
            <IconButton
              onClick={ props.hideView }
            >
              <ArrowBack />
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
              <DeleteOutline />
            </IconButton>
          </div>
        </div>
      </div>
      <Divider />
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
      <DatePicker
        className="mt-5"
        minDate={ moment().startOf('day').add(1, 'day') }
        onChange={ value => formik.setFieldValue('dueDate', value.startOf('day')) }
        inputFormat="MM/DD/yyyy"
        value={ formik.values.dueDate }
        renderInput={ params => <TextField fullWidth { ...params } /> }
      />
      <FormControl fullWidth className="mt-5">
        <InputLabel id="edit-funding-schedule-label">Funding Schedule</InputLabel>
        <Select
          data-testid="funding-schedule-selector"
          labelId="edit-funding-schedule-label"
          id="edit-funding-schedule"
          name="fundingScheduleId"
          value={ formik.values.fundingScheduleId }
          onChange={ formik.handleChange }
          disabled={ formik.isSubmitting }
        >
          { Array.from(fundingSchedules.values())
            .map(item => (
              <MenuItem
                key={ item.fundingScheduleId }
                value={ item.fundingScheduleId }
              >
                { item.name }
              </MenuItem>
            )) }
        </Select>
      </FormControl>
    </div>
  );
}
