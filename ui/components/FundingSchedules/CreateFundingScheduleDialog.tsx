import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { DatePicker } from '@mui/lab';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, TextField } from '@mui/material';
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
}

function CreateFundingScheduleDialog(): JSX.Element {
  const modal = useModal();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();
  const ref = useRef<HTMLDivElement>(null);

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
    });

    return createFundingSchedule(newFundingSchedule)
      .then(() => modal.remove())
      .finally(() => setSubmitting(false));
  }

  const initialValues: CreateFundingScheduleForm = {
    name: '',
    nextOccurrence: moment().add(1, 'day'),
    recurrenceRule: new Recurrence(),
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
          <Dialog open={ modal.visible } maxWidth="xs" ref={ ref }>
            <DialogTitle>
              Create A New Funding Schedule
            </DialogTitle>
            <DialogContent>
              <DialogContentText>
                Funding schedules tell monetr when you get paid. This way monetr can allocate funds towards your goals
                and expenses each time you get paid, in order to make sure you are covered when those expenses are due.
              </DialogContentText>
              <div className='grid sm:grid-cols-1 md:grid-cols-12 mt-5 gap-x-5 gap-y-5'>
                <div className='col-span-12'>
                  <span className='font-normal ml-3'>
                    What do you want to call your funding schedule?
                  </span>
                  <TextField
                    error={ touched.name && !!errors.name }
                    placeholder="Example: Payday..."
                    helperText={ (touched.name && errors.name) ? errors.name : ' ' }
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
                  <span className='font-normal ml-3'>
                    When do you get paid next?
                  </span>
                  <DatePicker
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
                  <span className='font-normal ml-3'>
                    How often do you get paid?
                  </span>
                  <RecurrenceSelect
                    menuRef={ ref.current }
                    disabled={ isSubmitting }
                    date={ values.nextOccurrence }
                    onChange={ value => setFieldValue('recurrenceRule', value) }
                  />
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
