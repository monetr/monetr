import { format, isThisYear } from 'date-fns';
import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { rrulestr } from 'rrule';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import Badge from '@monetr/interface/components/Badge';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import type Spending from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import Typography from '@monetr/interface/components/Typography';

export interface ExpenseItemProps {
  spending: Spending;
}

export default function ExpenseItem({ spending }: ExpenseItemProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: fundingSchedule } = useFundingSchedule(spending.fundingScheduleId);
  const rule = rrulestr(spending.ruleset!);

  const amountClass = mergeTailwind(
    {
      // Green if the expense is funded fully
      'text-green-500': spending.targetAmount === spending.currentAmount,
      // Blue if the expense is not funded fully but is not behind
      'text-blue-500': !spending.isBehind && spending.targetAmount > spending.currentAmount,
      // Red if the expense is not fully funded and is behind
      'text-red-500': spending.isBehind && spending.targetAmount > spending.currentAmount,
      // Yellow if the expense is over funded
      'text-yellow-400': spending.targetAmount < spending.currentAmount,
    },
    'text-end',
    'font-semibold',
  );

  const detailsPath = `/bank/${spending.bankAccountId}/expenses/${spending.spendingId}/details`;

  const dateString = isThisYear(spending.nextRecurrence!)
    ? format(spending.nextRecurrence!, 'MMM do')
    : format(spending.nextRecurrence!, 'MMM do, yyyy');

  return (
    <li className='group relative w-full px-1 md:px-2'>
      <Link
        className='absolute left-0 top-0 flex h-full w-full cursor-pointer md:hidden md:cursor-auto'
        to={detailsPath}
      />
      <div className='w-full flex rounded-lg group-hover:bg-zinc-600 gap-2 items-center px-2 py-1 cursor-pointer md:cursor-auto'>
        <div className='flex items-center flex-1 w-full md:w-1/2 gap-4 min-w-0 pr-1'>
          <MerchantIcon name={spending.name} />
          <div className='flex flex-col overflow-hidden min-w-0'>
            <div className='flex'>
              <Typography color='emphasis' weight='semibold' ellipsis>
                {spending.name}
              </Typography>
              <Badge size='xs' className='flex-none ml-1'>
                {dateString}
              </Badge>
            </div>
            {/* This block only shows on mobile screens */}
            <span className='hidden md:block text-zinc-200 font-sm text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              {capitalize(rule.toText())}
            </span>
            <span className='md:hidden text-zinc-200 text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              {locale.formatAmount(spending.nextContributionAmount, AmountType.Stored)} / {fundingSchedule?.name}
            </span>
          </div>
        </div>

        {/* This block only shows on desktop screens */}
        <div className='hidden md:flex w-1/2 overflow-hidden flex-1 min-w-0 items-center'>
          <span className='text-zinc-50/75 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
            {locale.formatAmount(spending.nextContributionAmount, AmountType.Stored)} / {fundingSchedule?.name}
          </span>
        </div>

        {/* This block only shows on mobile screens */}
        <div className='flex md:hidden shrink-0 items-center gap-2'>
          <div className='flex flex-col'>
            <span className={amountClass}>{locale.formatAmount(spending.currentAmount, AmountType.Stored)}</span>
            <hr className='w-full border-0 border-b-[thin] border-zinc-600' />
            <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
              {locale.formatAmount(spending.targetAmount, AmountType.Stored)}
            </span>
          </div>
          <ChevronRight className='text-zinc-600 group-hover:text-zinc-50 flex-none md:cursor-pointer' />
        </div>

        {/* This block only shows on desktops or larger screens */}
        <div className='hidden md:flex md:min-w-[12em] shrink-0 justify-end gap-2 items-center'>
          <div className='flex flex-col'>
            <div className='flex justify-end'>
              <span className={amountClass}>{locale.formatAmount(spending.currentAmount, AmountType.Stored)}</span>
              &nbsp;
              <span className='text-end text-zinc-500 group-hover:text-zinc-400 font-medium'>of</span>
              &nbsp;
              <span className='text-end text-zinc-400 group-hover:text-zinc-300 font-medium'>
                {locale.formatAmount(spending.targetAmount, AmountType.Stored)}
              </span>
            </div>
          </div>
          <ArrowLink to={detailsPath} />
        </div>
      </div>
    </li>
  );
}
