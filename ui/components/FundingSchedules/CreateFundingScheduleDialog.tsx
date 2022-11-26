import useIsMobile from 'hooks/useIsMobile';
import React, { Fragment, useRef, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Science } from '@mui/icons-material';
import { DatePicker } from '@mui/lab';
import { Button, Collapse, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControlLabel, InputAdornment, Switch, TextField } from '@mui/material';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import moment from 'moment';

import Recurrence from 'components/Recurrence/Recurrence';
import RecurrenceSelect from 'components/Recurrence/RecurrenceSelect';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useCreateFundingSchedule } from 'hooks/fundingSchedules';
import FundingSchedule from 'models/FundingSchedule';

interface CreateFundingScheduleForm {
  name: string;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;
}

function CreateFundingScheduleDialog(): JSX.Element {
  const modal = useModal();
  const isMobile = useIsMobile();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();
  const ref = useRef<HTMLDivElement>(null);

  const [showAdvanced, setShowAdvanced] = useState<boolean>(false);

  function validateInput(input: CreateFundingScheduleForm): FormikErrors<CreateFundingScheduleForm> {
    const errors: FormikErrors<CreateFundingScheduleForm> = {};
    if (input.name.trim().length === 0) {
      errors['name'] = 'Required to create a funding schedule.';
    }

    return errors;
  }

  async function submit(values: CreateFundingScheduleForm, { setSubmitting }: FormikHelpers<CreateFundingScheduleForm>): Promise<void> {
    setSubmitting(true);
    const newFundingSchedule = new FundingSchedule({
      bankAccountId: selectedBankAccountId,
      name: values.name,
      description: values.recurrenceRule.name,
      nextOccurrence: values.nextOccurrence.startOf('day'),
      rule: values.recurrenceRule.ruleString(),
      estimatedDeposit: values.estimatedDeposit && Math.ceil(values.estimatedDeposit * 100), // Convert to an integer.,
      excludeWeekends: values.excludeWeekends,
    });

    return createFundingSchedule(newFundingSchedule)
      .then(() => modal.remove())
      .finally(() => setSubmitting(false));
  }

  const initialValues: CreateFundingScheduleForm = {
    name: '',
    nextOccurrence: moment().add(1, 'day'),
    recurrenceRule: new Recurrence(),
    excludeWeekends: false,
    estimatedDeposit: null,
  };

  return (
    <Formik
      initialValues={ initialValues }
      validate={ validateInput }
      onSubmit={ submit }
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
          <Dialog open={ modal.visible } maxWidth="xs" ref={ ref } fullScreen={ isMobile }>
            <DialogTitle>
              Create A New Funding Schedule
            </DialogTitle>
            <DialogContent className='pb-0 pt-0'>
              <DialogContentText>
                Funding schedules tell monetr when you get paid. This way monetr can allocate funds towards your goals
                and expenses each time you get paid, in order to make sure you are covered when those expenses are due.
              </DialogContentText>
              <div className='grid sm:grid-cols-1 md:grid-cols-12 mt-5 mb-6 gap-x-5 gap-y-5'>
                <div className='col-span-12'>
                  <TextField
                    label="What do you want to call your funding schedule?"
                    error={ touched.name && values.name && !!errors.name }
                    placeholder="Example: Payday..."
                    helperText={ (touched.name && values.name && errors.name) ? errors.name : ' ' }
                    name="name"
                    className="w-full"
                    onChange={ handleChange }
                    onBlur={ handleBlur }
                    value={ values.name }
                    disabled={ isSubmitting }
                    required
                  />
                </div>
                <div className='col-span-12'>
                  <DatePicker
                    label="When do you get paid next?"
                    disabled={ isSubmitting }
                    minDate={ moment().startOf('day').add(1, 'day') }
                    onChange={ value => setFieldValue('nextOccurrence', value.startOf('day')) }
                    inputFormat="MM/DD/yyyy"
                    value={ values.nextOccurrence }
                    renderInput={ params => (
                      <TextField label="When do you get paid next?"  fullWidth { ...params } />
                    ) }
                  />
                </div>
                <div className='col-span-12'>
                  <RecurrenceSelect
                    label="How often do you get paid?"
                    menuRef={ ref.current }
                    disabled={ isSubmitting }
                    date={ values.nextOccurrence }
                    onChange={ value => setFieldValue('recurrenceRule', value) }
                  />
                </div>
                <div className='col-span-12'>
                  <FormControlLabel
                    control={ <Switch value={ showAdvanced } onChange={ (_, checked) => setShowAdvanced(checked) } /> }
                    label="Show Advanced Options"
                  />
                  <Collapse in={ showAdvanced }>
                    <FormControlLabel
                      control={
                        <Switch
                          value={ values.excludeWeekends }
                          onChange={ (_, checked) => setFieldValue('excludeWeekends', checked) }
                        />
                      }
                      label={
                        <Fragment>
                          <span>Exclude Weekends</span> <Science className="mb-1 fill-gray-600" />
                        </Fragment>
                      }
                    />
                    <div className='col-span-12 pt-5'>
                      <TextField
                        label="About how much do you get paid each time?"
                        error={ touched.estimatedDeposit && !!errors.estimatedDeposit }
                        placeholder="Example: $1200"
                        helperText={ (touched.estimatedDeposit && errors.estimatedDeposit) ? errors.estimatedDeposit : ' ' }
                        name="estimatedDeposit"
                        className="w-full"
                        onChange={ handleChange }
                        onBlur={ handleBlur }
                        value={ values.estimatedDeposit }
                        disabled={ isSubmitting }
                        type="number"
                        InputProps={ {
                          startAdornment: <InputAdornment position="start">$</InputAdornment>,
                          inputProps: { min: 0 },
                        } }
                      />
                    </div>
                  </Collapse>
                </div>
              </div>
            </DialogContent>
            <DialogActions>
              <Button
                color="secondary"
                disabled={ isSubmitting }
                onClick={ modal.remove }
              >
                Cancel
              </Button>
              <Button
                disabled={ isSubmitting }
                onClick={ submitForm }
                color="primary"
                type="submit"
              >
                Create
              </Button>
            </DialogActions>
          </Dialog>
        </form>
      ) }
    </Formik>
  );
}

const createFundingScheduleModal = NiceModal.create(CreateFundingScheduleDialog);

export default createFundingScheduleModal;

export function showCreateFundingScheduleDialog(): void {
  NiceModal.show(createFundingScheduleModal);
}
