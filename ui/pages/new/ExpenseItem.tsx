import React from 'react';
import { KeyboardArrowRight } from '@mui/icons-material';
import { Avatar } from '@mui/material';
import clsx from 'clsx';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import Spending from 'models/Spending';
import { rrulestr } from 'rrule';

export interface ExpenseItemProps {
  spending: Spending;
}

export default function ExpenseItem({ spending }: ExpenseItemProps): JSX.Element {
  const fundingSchedule = useFundingSchedule(spending.fundingScheduleId);

  const letter = spending.name.toUpperCase().charAt(0);
  const rule = rrulestr(spending.recurrenceRule);

  const amountClass = clsx([
    'text-end',
    'font-semibold',
  ], {
    'text-green-500': spending.targetAmount === spending.currentAmount,
    'text-blue-500': spending.targetAmount !== spending.currentAmount,
  },);

  return (
    <li className='w-full px-2'>
      <div className='w-full flex rounded-lg hover:bg-zinc-600 group gap-2 items-center px-2 py-1'>
        <div className='flex items-center flex-grow w-full gap-4 min-w-0'>
          <Avatar className='bg-zinc-800 h-10 w-10'>
            { letter }
          </Avatar>
          <div className='flex flex-col justify-center overflow-hidden'>
            <span
              className='text-zinc-50 font-semibold text-base overflow-hidden text-ellipsis whitespace-nowrap min-w-0'
            >
              { spending.name }
            </span>
            <span className='text-zinc-200 font-medium text-sm overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { spending.getNextContributionAmountString() } / { fundingSchedule?.name }
            </span>
          </div>
        </div>
        <div className='flex md:min-w-[8em] shrink-0 justify-end gap-2 items-center'>
          <div className='flex flex-col'>
            <div className='flex justify-end'>
              <span className={ amountClass }>
                { spending.getCurrentAmountString() }
              </span>
              &nbsp;
              <span className='text-end text-zinc-500 font-medium'>
                of
              </span>
              &nbsp;
              <span className='text-end text-zinc-400 font-medium'>
                { spending.getTargetAmountString() }
              </span>
            </div>
            <span className='text-zinc-200 font-medium text-sm overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { rule.toText() }
            </span>
          </div>
          <KeyboardArrowRight className='text-zinc-600 group-hover:text-zinc-50 flex-none md:cursor-pointer' />
        </div>
      </div>
    </li>
  );
}
