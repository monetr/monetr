/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { ArrowBackOutlined, ExitToAppOutlined, HeartBroken, MenuOutlined, PriceCheckOutlined, SaveOutlined } from '@mui/icons-material';

import MSelect from 'components/MSelect';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useFundingSchedulesSink } from 'hooks/fundingSchedules';
import { useSpending } from 'hooks/spending';
import MerchantIcon from 'pages/new/MerchantIcon';
import { MBaseButton } from 'components/MButton';
import MDivider from 'components/MDivider';
import { Stats } from 'fs-extra';

export interface ExpenseDetailsProps {
  spendingId?: number;
}

export default function ExpenseDetails(props: ExpenseDetailsProps): JSX.Element {
  const spending = useSpending(props.spendingId);
  const { result: fundingSchedules } = useFundingSchedulesSink();

  if (!props.spendingId) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          There wasn't an expense specified...
        </MSpan>
      </div>
    );
  }

  const options = Array.from(fundingSchedules.values())
    .map(fundingSchedule => ({
      label: fundingSchedule.name,
      value: fundingSchedule.fundingScheduleId,
    }));
  const currentOption = options.find(option => option.value === spending?.fundingScheduleId) || undefined;

  return (
    <Fragment>
      <div className='w-full h-auto md:h-12 flex flex-col md:flex-row md:items-center px-4 gap-4 md:justify-between'>
        <div className='flex items-center gap-2 mt-2 md:mt-0'>
          <MenuOutlined className='visible lg:hidden dark:text-dark-monetr-content-emphasis cursor-pointer mr-2' />
          <span className='text-2xl dark:text-dark-monetr-content-subtle font-bold'>
            <PriceCheckOutlined />
          </span>
          <span className='text-2xl hidden md:block dark:text-dark-monetr-content-subtle dark:hover:text-dark-monetr-content-emphasis font-bold cursor-pointer'>
            Expenses
          </span>
          <span className='text-2xl hidden md:block dark:text-dark-monetr-content-subtle font-bold'>
            /
          </span>
          <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold'>
            { spending?.name }
          </span>
        </div>
        <div className='flex gap-2'>
          <MBaseButton color='cancel' className='gap-1 py-1 px-2'>
            <ArrowBackOutlined />
            Cancel
          </MBaseButton>
          <MBaseButton color='primary' className='gap-1 py-1 px-2'>
            <SaveOutlined />
            Save Changes
          </MBaseButton>
        </div>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full h-full gap-4'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <div className='w-full flex justify-center mb-2'>
              <MerchantIcon name={ spending?.name } />
            </div>
            <MTextField
              label='Expense name'
              name='name'
              value={ spending?.name }
              className='w-full'
            />
            <MTextField
              label='Amount'
              name='amount'
              prefix='$'
              type='number'
              value={ (spending?.targetAmount / 100).toFixed(2) }
              className='w-full'
            />
            <MTextField
              label='Next Occurrence'
              name='nextRecurrence'
              type='date'
              value={ spending?.nextRecurrence.format('YYYY-MM-DD') }
              className='w-full'
            />
            <MSelect
              label='Funding Schedule'
              name='fundingScheduleId'
              options={ options }
              value={ currentOption }
              className='w-full'
            />
            <MSelect
              label='Spending Frequency'
              name='recurrenceRule'
              placeholder='Select a spending frequency...'
              options={ [] }
              className='w-full'
            />
            <MDivider className='w-1/2' />
            <MSpan className='my-2'>Stats</MSpan>
            <div className='w-full'>
              <MSpan>Estimated Regular Contribution Amount:</MSpan>
              &nbsp;
              <MSpan>$140.00</MSpan>


            </div>
          </div>
          <div className='w-full md:w-1/2'>
            <MSpan className='text-xl'>
              Expense Timeline
            </MSpan>
          </div>
        </div>
      </div>
    </Fragment>
  );
}
