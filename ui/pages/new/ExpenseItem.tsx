/* eslint-disable max-len */
import React from 'react';
import { KeyboardArrowRight } from '@mui/icons-material';
import { Avatar } from '@mui/material';
import clsx from 'clsx';

import { useFundingSchedule } from 'hooks/fundingSchedules';
import { useIconSearch } from 'hooks/useIconSearch';
import Spending from 'models/Spending';
import { rrulestr } from 'rrule';

export interface ExpenseItemProps {
  spending: Spending;
}

export default function ExpenseItem({ spending }: ExpenseItemProps): JSX.Element {
  const fundingSchedule = useFundingSchedule(spending.fundingScheduleId);
  const icon = useIconSearch(spending.name);
  const IconContent = () => {
    if (icon?.svg) {
      // It is possible for colors to be missing for a given icon. When this happens just fall back to a black color.
      const colorStyles = icon?.colors?.length > 0 ?
        { backgroundColor: `#${icon.colors[0]}` } :
        { backgroundColor: '#000000' };

      const styles = {
        WebkitMaskImage: `url(data:image/svg+xml;base64,${icon.svg})`,
        WebkitMaskRepeat: 'no-repeat',
        height: '30px',
        width: '30px',
        ...colorStyles,
      };

      return (
        <div className='bg-white flex items-center justify-center h-10 w-10 rounded-full'>
          <div style={ styles } />
        </div>
      );
    }

    // If we have no icon to work with then create an avatar with the first character of the transaction name.
    const letter = spending.name.toUpperCase().charAt(0);
    return (
      <Avatar className='bg-zinc-800 h-10 w-10 text-zinc-200'>
        {letter}
      </Avatar>
    );
  };

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
      <div className='w-full flex rounded-lg hover:bg-zinc-600 group gap-2 items-center px-2 py-1 cursor-pointer md:cursor-auto'>
        <div className='flex items-center flex-1 w-full md:w-1/2 gap-4 min-w-0 pr-1'>
          <IconContent />
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
          <span className='flex-none text-zinc-50/75 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
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
          <KeyboardArrowRight className='text-zinc-600 group-hover:text-zinc-50 flex-none md:cursor-pointer' />
        </div>
      </div>
    </li>
  );
}
