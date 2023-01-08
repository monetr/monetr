import { TextField } from '@mui/material';
import MForm from 'components/MForm';
import MTextField from 'components/MTextField';
import Recurrence from 'components/Recurrence/Recurrence';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import { useFundingSchedule } from 'hooks/fundingSchedules';
import React from 'react';
import { useParams } from 'react-router-dom';
import { RRule } from 'rrule';

interface FundingScheduleEditForm {
  name: string;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;
}

export default function FundingEditPage(): JSX.Element {
  const params = useParams();
  const fundingScheduleId = params['fundingScheduleId'];
  // TODO Make the parsing of the parameter safer and more graceful.
  const fundingSchedule = useFundingSchedule(+fundingScheduleId);

  if (!fundingSchedule) {
    // TODO Implement a proper loading state here?
    return (
      <div className='minus-nav'>
        <div className='w-full view-area bg-white px-2 py-2'>
          <div>
            <div>
              <span className='font-medium text-2xl w-full'>
                Funding Schedule Not Found
              </span>
              <p className='md:w-1/2 w-full font-normal text-slate-600'>
                The funding schedule you were looking to edit cannot be found.
              </p>
            </div>
          </div>
        </div>
      </div>
    )
  }

  const initialValues: FundingScheduleEditForm = {
    name: fundingSchedule.name,
    nextOccurrence: fundingSchedule.nextOccurrence,
    recurrenceRule: new Recurrence({
      rule: RRule.fromText(`RRULE: ${ fundingSchedule.rule}`),
    }),
    excludeWeekends: false,
    estimatedDeposit: null,
  };

  function validateInput(input: FundingScheduleEditForm): FormikErrors<FundingScheduleEditForm> {
    return {};
  }

  async function submit(values: FundingScheduleEditForm, helpers: FormikHelpers<FundingScheduleEditForm>) {
    return null;
  }

  return (
    <div className='minus-nav'>
      <div className='w-full view-area bg-white px-2 py-2'>
        <div>
          <div>
            <span className='font-medium text-2xl w-full'>
              Edit Funding Schedule
            </span>
            <p className='md:w-1/2 w-full font-normal text-slate-600'>
              Here you can edit your funding schedule, update when you'll get paid next, how frequently you get paid, or
              how much you get paid. As well as some other tweaks to help make your budgeting experience even easier.
            </p>
          </div>
          <div className='py-5'>
            <Formik
              initialValues={ initialValues }
              validate={ validateInput }
              onSubmit={ submit }
            >
              <MForm>
                <MTextField
                  name="name"
                  label="Funding Schedule Name"
                  required
                />
              </MForm>
            </Formik>
          </div>
        </div>
      </div>
    </div>
  )
}
