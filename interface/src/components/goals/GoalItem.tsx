/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { useNavigate } from 'react-router-dom';
import { KeyboardArrowRight } from '@mui/icons-material';

import MBadge from '@monetr/interface/components/MBadge';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import { useFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import Spending from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface GoalItemProps {
  spending: Spending;
}

export default function GoalItem({ spending }: GoalItemProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: fundingSchedule } = useFundingSchedule(spending.fundingScheduleId);
  const navigate = useNavigate();


  function openDetails() {
    navigate(`/bank/${spending.bankAccountId}/goals/${spending.spendingId}/details`);
  }

  // By default the contribution string should simply be the amount that will be added to this goal per funding schedule
  // it is associated with.
  let contributionString = `${ locale.formatAmount(spending.nextContributionAmount, AmountType.Stored)} / ${ fundingSchedule?.name }`;
  // But if the goal is no longer in progress (it is complete). Then indicate that.
  if (!spending.getGoalIsInProgress()) {
    contributionString = 'Complete';
  } else if (spending.isPaused) { // Or if the goal is just paused.
    contributionString = 'Paused';
  }

  return (
    <li className='group relative w-full px-1 md:px-2'>
      <div
        className='absolute left-0 top-0 flex h-full w-full cursor-pointer md:hidden md:cursor-auto'
        onClick={ openDetails }
      />
      <div className='w-full flex rounded-lg group-hover:bg-zinc-600 gap-2 md:gap-4 items-center px-2 py-1 cursor-pointer md:cursor-auto'>
        <MerchantIcon name={ spending.name } />
        <div className='w-full flex flex-col min-w-0'>
          <div className='w-full flex gap-2 items-center min-w-0 justify-between md:justify-normal'>
            <div className='flex items-center md:w-1/2 gap-4 min-w-0 pr-1'>
              <div className='flex flex-col overflow-hidden min-w-0'>
                <span className='block w-full text-zinc-50 font-semibold text-base overflow-hidden whitespace-nowrap truncate'>
                  { spending.name }
                  <span className='md:hidden text-zinc-200 font-sm text-sm overflow-hidden whitespace-nowrap truncate'>
                    &nbsp;• { spending.getNextOccurrenceString() }
                  </span>
                </span>
                <span className='hidden md:block text-zinc-200 font-sm text-sm overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
                  { spending.getNextOccurrenceString() }
                </span>
                <span className='md:hidden text-zinc-200 text-sm overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
                  { contributionString }
                </span>
              </div>
            </div>
            <div className='hidden md:flex w-1/2 overflow-hidden flex-1 min-w-0 items-center'>
              <span className='text-zinc-50/75 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
                { contributionString }
              </span>
            </div>
            <GoalAmount spending={ spending } />
          </div>
          <GoalProgressBar spending={ spending } />
        </div>
        <KeyboardArrowRight
          className='dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis md:cursor-pointer'
          onClick={ openDetails }
        />
      </div>
    </li>
  );
}

interface GoalProps {
  spending: Spending;
}

function GoalAmount({ spending }: GoalProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const user = useAuthentication();
  const amountClass = mergeTailwind(
    {
      'text-green-500': spending.targetAmount <= spending.currentAmount,
      'text-blue-500': spending.targetAmount !== spending.currentAmount,
    },
    'text-end',
    'font-semibold',
  );

  const currentAmountString = locale.formatAmount(spending.currentAmount, AmountType.Stored);
  const targetAmountString = locale.formatAmount(spending.targetAmount, AmountType.Stored);

  if (spending.getGoalIsInProgress()) {
    return (
      <Fragment>
        <div className='flex md:hidden shrink-0 items-center gap-2'>
          <div className='flex flex-col'>
            <span className={ amountClass }>
              { currentAmountString }
            </span>
            <hr className='w-full border-0 border-b-[thin] border-zinc-600' />
            <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
              { targetAmountString }
            </span>
          </div>
        </div>
        <div className='hidden md:flex md:min-w-[12em] shrink-0 justify-end gap-2 items-center'>
          <div className='flex flex-col'>
            <div className='flex justify-end'>
              <span className={ amountClass }>
                { currentAmountString }
              </span>
              &nbsp;
              <span className='text-end text-zinc-500 group-hover:text-zinc-400 font-medium'>
                of
              </span>
              &nbsp;
              <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                { targetAmountString }
              </span>
            </div>
          </div>
        </div>
      </Fragment>
    );
  }

  return (
    <div className='flex md:min-w-[12em] shrink-0 justify-end gap-2 items-center'>
      <MBadge className='w-fit justify-end dark:bg-green-600' weight='medium'>
        { locale.formatAmount(spending.currentAmount, AmountType.Stored) }
      </MBadge>
    </div>
  );
}

function GoalProgressBar({ spending }: GoalProps): JSX.Element {
  const { usedAmount, currentAmount, targetAmount } = spending;
  const usedProgress = ((Math.min(usedAmount, targetAmount) / targetAmount) * 100).toFixed(0);
  const allocatedProgress = ((Math.min(currentAmount + usedAmount, targetAmount) / targetAmount) * 100).toFixed(0);
  return (
    <div className='w-full bg-gray-200 rounded-full h-1.5 my-2 dark:bg-gray-700 relative'>
      <div className='absolute top-0 bg-green-600 h-1.5 rounded-full dark:bg-green-600' style={ { width: `${allocatedProgress}%` } }></div>
      <div className='absolute top-0 bg-blue-600 h-1.5 rounded-full dark:bg-blue-600' style={ { width: `${usedProgress}%` } }></div>
    </div>
  );
}
