import { isThisYear } from 'date-fns';

import { useLocale } from '@monetr/interface/hooks/useLocale';

interface TransactionDateItemProps {
  date: Date;
}

export default function TransactionDateItem({ date }: TransactionDateItemProps): JSX.Element {
  const { data: locale, isLoading } = useLocale();

  if (isLoading) {
    return (
      <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90% mr-4'>
        <span className='dark:text-dark-monetr-content-subtle font-semibold text-base z-10 px-3 md:px-4'>
          Loading...
        </span>
      </li>
    );
  }

  const dateString = new Intl.DateTimeFormat(locale.code, {
    month: 'long',
    day: 'numeric',
    // Only include the year if it is a different year than the current year.
    year: isThisYear(date) ? undefined : 'numeric',
  }).format(date);

  return (
    <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90% mr-4'>
      <span className='dark:text-dark-monetr-content-subtle font-semibold text-base z-10 px-3 md:px-4'>
        {dateString}
      </span>
    </li>
  );
}
