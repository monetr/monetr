/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { ArrowBackOutlined, HeartBroken, MenuOutlined, PriceCheckOutlined, SaveOutlined } from '@mui/icons-material';
import { Formik } from 'formik';

import { MBaseButton } from 'components/MButton';
import MDivider from 'components/MDivider';
import MForm from 'components/MForm';
import MSelect from 'components/MSelect';
import MSpan from 'components/MSpan';
import MTextField from 'components/MTextField';
import { useFundingSchedulesSink } from 'hooks/fundingSchedules';
import { useSpending } from 'hooks/spending';
import MerchantIcon from 'pages/new/MerchantIcon';

export interface ExpenseDetailsProps {
  spendingId?: number;
}

interface ExpenseValues {
  name: string;
  amount: number;
  nextRecurrence: moment.Moment;
  fundingScheduleId: number;
  recurrenceRule: string;
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

  if (!spending) {
    return null;
  }

  function submit() {

  }

  const initialValues: ExpenseValues = {
    name: spending.name,
    amount: +(spending.targetAmount / 100).toFixed(2),
    nextRecurrence: spending.nextRecurrence,
    fundingScheduleId: spending.fundingScheduleId,
    recurrenceRule: spending.recurrenceRule,
  };

  const options = Array.from(fundingSchedules.values())
    .map(fundingSchedule => ({
      label: fundingSchedule.name,
      value: fundingSchedule.fundingScheduleId,
    }));

  return (
    <Formik
      initialValues={ initialValues }
      onSubmit={ submit }
    >
      <MForm className='flex w-full h-full flex-col'>
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
          <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
            <div className='w-full md:w-1/2 flex flex-col items-center'>
              <div className='w-full flex justify-center mb-2'>
                <MerchantIcon name={ spending?.name } />
              </div>
              <MTextField
                label='Expense'
                name='name'
                value={ spending?.name }
                className='w-full'
              />
              <MTextField
                label='Amount'
                name='amount'
                prefix='$'
                type='number'
                className='w-full'
              />
              <MTextField
                label='Next Occurrence'
                name='nextRecurrence'
                type='date'
                className='w-full'
              />
              <MSelect
                label='Funding Schedule'
                name='fundingScheduleId'
                options={ options }
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
              <MSpan className='text-xl my-2'>
              Stats
              </MSpan>
              <div className='w-full'>
                <MSpan>Estimated Regular Contribution Amount:</MSpan>
              &nbsp;
                <MSpan>$140.00</MSpan>


              </div>
            </div>
            <MDivider className='block md:hidden w-1/2' />
            <div className='w-full md:w-1/2 flex flex-col gap-2'>
              <MSpan className='text-xl my-2'>
              Expense Timeline
              </MSpan>
              <ol className="relative border-l border-zinc-200 dark:border-zinc-700">
                <li className="mb-10 ml-4">
                  <div className="absolute w-3 h-3 bg-zinc-200 rounded-full mt-1.5 -left-1.5 border border-white dark:border-zinc-900 dark:bg-zinc-700"></div>
                  <time className="mb-1 text-sm font-normal leading-none text-zinc-400 dark:text-zinc-500">February 2022</time>
                  <h3 className="text-lg font-semibold text-zinc-900 dark:text-white">Application UI code in Tailwind CSS</h3>
                  <p className="mb-4 text-base font-normal text-zinc-500 dark:text-zinc-400">Get access to over 20+ pages including a dashboard layout, charts, kanban board, calendar, and pre-order E-commerce & Marketing pages.</p>
                  <a href="#" className="inline-flex items-center px-4 py-2 text-sm font-medium text-zinc-900 bg-white border border-zinc-200 rounded-lg hover:bg-zinc-100 hover:text-blue-700 focus:z-10 focus:ring-4 focus:outline-none focus:ring-zinc-200 focus:text-blue-700 dark:bg-zinc-800 dark:text-zinc-400 dark:border-zinc-600 dark:hover:text-white dark:hover:bg-zinc-700 dark:focus:ring-zinc-700">Learn more <svg className="w-3 h-3 ml-2" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
                    <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 5h12m0 0L9 1m4 4L9 9" />
                  </svg></a>
                </li>
                <li className="mb-10 ml-4">
                  <div className="absolute w-3 h-3 bg-zinc-200 rounded-full mt-1.5 -left-1.5 border border-white dark:border-zinc-900 dark:bg-zinc-700"></div>
                  <time className="mb-1 text-sm font-normal leading-none text-zinc-400 dark:text-zinc-500">March 2022</time>
                  <h3 className="text-lg font-semibold text-zinc-900 dark:text-white">Marketing UI design in Figma</h3>
                  <p className="text-base font-normal text-zinc-500 dark:text-zinc-400">All of the pages and components are first designed in Figma and we keep a parity between the two versions even as we update the project.</p>
                </li>
                <li className="ml-4">
                  <div className="absolute w-3 h-3 bg-zinc-200 rounded-full mt-1.5 -left-1.5 border border-white dark:border-zinc-900 dark:bg-zinc-700"></div>
                  <time className="mb-1 text-sm font-normal leading-none text-zinc-400 dark:text-zinc-500">April 2022</time>
                  <h3 className="text-lg font-semibold text-zinc-900 dark:text-white">E-Commerce UI code in Tailwind CSS</h3>
                  <p className="text-base font-normal text-zinc-500 dark:text-zinc-400">Get started with dozens of web components and interactive elements built on top of Tailwind CSS.</p>
                </li>
              </ol>
            </div>
          </div>
        </div>
      </MForm>
    </Formik>
  );
}
