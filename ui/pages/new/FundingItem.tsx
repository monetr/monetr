/* eslint-disable max-len */
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { KeyboardArrowRight } from '@mui/icons-material';
import { Avatar } from '@mui/material';

import { useNextFundingForecast } from 'hooks/forecast';
import FundingSchedule from 'models/FundingSchedule';
import { rrulestr } from 'rrule';
import formatAmount from 'util/formatAmount';

export interface FundingItemProps {
  funding: FundingSchedule;
}

export default function FundingItem(props: FundingItemProps): JSX.Element {
  const navigate = useNavigate();
  const { funding } = props;
  const contributionForecast = useNextFundingForecast(funding.fundingScheduleId);
  const rule = rrulestr(funding.rule);
  const letter = funding.name.toUpperCase().charAt(0) || '?';

  function openDetails() {
    navigate(`/bank/${funding.bankAccountId}/funding/${funding.fundingScheduleId}/details`);
  }

  return (
    <li className='w-full px-1 md:px-2'>
      <div className='flex rounded-lg hover:bg-zinc-600 gap-1 md:gap-4 group px-2 py-1 h-full cursor-pointer md:cursor-auto'>
        <div className='w-full md:w-1/2 flex flex-row gap-4 items-center flex-1 min-w-0'>
          <Avatar className='bg-zinc-800 h-10 w-10 text-zinc-200'>
            { letter }
          </Avatar>
          <div className='flex flex-col overflow-hidden min-w-0'>
            <span className='text-zinc-50 font-semibold text-base w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { funding.name }
            </span>
            <span className='dark:text-dark-monetr-content font-medium text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              { rule.toText() }
            </span>
          </div>
        </div>
        <div className='flex md:min-w-[14em] shrink-0 justify-end gap-2 items-center'>
          <div className='flex flex-col'>
            <div className='flex justify-end'>
              <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                Estimated Contribution
              </span>
              &nbsp;
              <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                { formatAmount(contributionForecast.data) }
              </span>
            </div>
          </div>
          <KeyboardArrowRight
            className='dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis flex-none md:cursor-pointer'
            onClick={ openDetails }
          />
        </div>
      </div>
    </li>
  );
}
