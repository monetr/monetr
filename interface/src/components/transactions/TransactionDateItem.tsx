/* eslint-disable max-len */

import { format, isThisYear } from 'date-fns';

interface TransactionDateItemProps {
  date: Date;
}

export default function TransactionDateItem({ date }: TransactionDateItemProps): JSX.Element {
  const dateString = isThisYear(date) ? format(date, 'MMMM do') : format(date, 'MMMM do, yyyy');

  return (
    <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90% mr-4'>
      <span className='dark:text-dark-monetr-content-subtle font-semibold text-base z-10 px-3 md:px-4'>
        {dateString}
      </span>
    </li>
  );
}
