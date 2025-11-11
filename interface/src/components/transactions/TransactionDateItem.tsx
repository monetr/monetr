import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

interface TransactionDateItemProps {
  date: Date;
}

export default function TransactionDateItem({ date }: TransactionDateItemProps): JSX.Element {
  const { inTimezone } = useTimezone();
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

  const dateString = formatDate(date, inTimezone, locale, DateLength.Long);
  return (
    <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90% mr-4'>
      <span className='dark:text-dark-monetr-content-subtle font-semibold text-base z-10 px-3 md:px-4'>
        {dateString}
      </span>
    </li>
  );
}
