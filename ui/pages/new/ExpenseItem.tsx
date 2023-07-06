/* eslint-disable max-len */
import React from 'react';
import { KeyboardArrowRight } from '@mui/icons-material';

import MerchantIcon from './MerchantIcon';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import Spending from 'models/Spending';
import { rrulestr } from 'rrule';
import mergeTailwind from 'util/mergeTailwind';
import { useNavigate } from 'react-router-dom';

export interface ExpenseItemProps {
  spending: Spending;
}

export default function ExpenseItem({ spending }: ExpenseItemProps): JSX.Element {
  const fundingSchedule = useFundingSchedule(spending.fundingScheduleId);
  const navigate = useNavigate();
  const rule = rrulestr(spending.recurrenceRule);

  const amountClass = mergeTailwind(
    {
      'text-green-500': spending.targetAmount === spending.currentAmount,
      'text-blue-500': spending.targetAmount !== spending.currentAmount,
    },
    'text-end',
    'font-semibold',
  );

  function openDetails() {
    navigate(`/expenses/${spending.spendingId}/details`);
  }

  return (
    <li className='w-full px-2'>
      <div className='w-full flex rounded-lg hover:bg-zinc-600 group gap-2 items-center px-2 py-1 cursor-pointer md:cursor-auto'>
        <div className='flex items-center flex-1 w-full md:w-1/2 gap-4 min-w-0 pr-1'>
          <MerchantIcon name={ spending.name } />
          <div className='flex flex-col overflow-hidden min-w-0'>
            <span className='text-zinc-50 font-semibold text-base w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { spending.name }
            </span>
            <span className='hidden md:block text-zinc-200 font-sm text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { rule.toText() }
            </span>
            <span className='md:hidden text-zinc-200 text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { spending.getNextContributionAmountString() } / { fundingSchedule?.name }
            </span>
          </div>
        </div>
        <div className='hidden md:flex w-1/2 overflow-hidden flex-1 min-w-0 items-center'>
          <span className='text-zinc-50/75 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
            { spending.getNextContributionAmountString() } / { fundingSchedule?.name }
          </span>
        </div>
        <div className='flex md:hidden shrink-0 items-center gap-2'>
          <div className='flex flex-col'>
            <span className={ amountClass }>
              { spending.getCurrentAmountString() }
            </span>
            <hr className='w-full border-0 border-b-[thin] border-zinc-600' />
            <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
              { spending.getTargetAmountString() }
            </span>
          </div>
          <KeyboardArrowRight className='text-zinc-600 group-hover:text-zinc-50 flex-none md:cursor-pointer' />
        </div>
        <div className='hidden md:flex md:min-w-[12em] shrink-0 justify-end gap-2 items-center'>
          <div className='flex flex-col'>
            <div className='flex justify-end'>
              <span className={ amountClass }>
                { spending.getCurrentAmountString() }
              </span>
              &nbsp;
              <span className='text-end text-zinc-500 group-hover:text-zinc-400 font-medium'>
                of
              </span>
              &nbsp;
              <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                { spending.getTargetAmountString() }
              </span>
            </div>
          </div>
          <KeyboardArrowRight className='text-zinc-600 group-hover:text-zinc-50 flex-none md:cursor-pointer' onClick={ openDetails } />
        </div>
      </div>
    </li>
  );
}
