/* eslint-disable max-len */
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { KeyboardArrowRight } from '@mui/icons-material';
import { Avatar } from '@mui/material';

import MSpan from 'components/MSpan';
import { format, isThisYear } from 'date-fns';
import { useNextFundingForecast } from 'hooks/forecast';
import FundingSchedule from 'models/FundingSchedule';
import { rrulestr } from 'rrule';
import capitalize from 'util/capitalize';
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

  const ruleDescription = capitalize(rule.toText());

  const next = funding.nextOccurrence;
  const dateFormatString = isThisYear(next) ? 'EEEE LLLL do' : 'EEEE LLLL do, yyyy';
  // TODO look into format distance.
  const nextOccurrenceString = `${format(next, dateFormatString)}`;

  function openDetails() {
    navigate(`/bank/${funding.bankAccountId}/funding/${funding.fundingScheduleId}/details`);
  }

  return (
    <li className='group relative w-full px-1 md:px-2'>
      <div
        className='absolute left-0 top-0 flex h-full w-full cursor-pointer md:hidden md:cursor-auto'
        onClick={ openDetails }
      />
      <div className='flex items-center rounded-lg group-hover:bg-zinc-600 gap-1 md:gap-4 px-2 py-1 h-full cursor-pointer md:cursor-auto'>
        <Avatar className='bg-zinc-800 h-10 w-10 text-zinc-200'>
          { letter }
        </Avatar>
        <div className='w-full md:w-1/2 flex flex-col flex-1 min-w-0 overflow-hidden'>
          <MSpan weight='semibold' color='emphasis' ellipsis>
            { funding.name }
          </MSpan>
          <MSpan size='sm' weight='medium' ellipsis>
            { ruleDescription }
          </MSpan>
          <MSpan size='sm' weight='medium' ellipsis>
            { nextOccurrenceString }
          </MSpan>
        </div>
        <div className='flex md:min-w-[14em] shrink-0 justify-end gap-2 items-center'>
          <div className='flex flex-col'>
            <div className='flex justify-end'>
              <span className='hidden sm:block text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                Estimated Contribution
              </span>
              <span className='block sm:hidden text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                Est.
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
