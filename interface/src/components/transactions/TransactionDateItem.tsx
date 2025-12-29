import { useMemo } from 'react';

import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

interface TransactionDateItemProps {
  date: Date;
}

export default function TransactionDateItem({ date }: TransactionDateItemProps): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale, isLoading } = useLocale();

  const dateString = useMemo(
    () => (isLoading ? 'Loading...' : formatDate(date, inTimezone, locale, DateLength.Long)),
    [locale, date, isLoading, inTimezone],
  );

  // Version with the sticky headers. We are removing this for now to test full body scrolling on iOS
  // return (
  //   <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90% mr-4'>
  //     <Typography className='z-10 px-3 md:px-4' color='subtle' weight='semibold'>
  //       {dateString}
  //     </Typography>
  //   </li>
  // );

  return (
    <li className='h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90% mr-4'>
      <Typography className='z-10 px-3 md:px-4' color='subtle' weight='semibold'>
        {dateString}
      </Typography>
    </li>
  );
}
