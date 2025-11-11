import { Link } from 'react-router-dom';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransaction } from '@monetr/interface/hooks/useTransaction';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

export interface SimilarTransactionItemProps {
  transactionId: string;
  /**
   * disableNavigate will remove the arrow link or the click-ability of the similar transaction item.
   */
  disableNavigate?: boolean;
}

export default function SimilarTransactionItem(props: SimilarTransactionItemProps): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: transaction, isLoading, isError } = useTransaction(props.transactionId);
  const { data: locale, isLoading: localeIsLoading } = useLocale();

  if (isLoading || localeIsLoading) {
    return (
      <li className='group relative w-full px-1 md:px-2'>
        <div className='group animate-pulse flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4'>
          <div className='flex w-full min-w-0 flex-1 flex-row items-center gap-4 md:w-1/2'>
            <div className='h-10 w-10 rounded-full dark:bg-dark-monetr-background-subtle' />
            <div className='flex min-w-0 grow flex-col overflow-hidden'>
              <div className='w-full rounded-xl h-4 my-1 dark:bg-dark-monetr-background-subtle' />
              <div className='w-1/2 rounded-xl h-3 my-1 dark:bg-dark-monetr-background-subtle opacity-70' />
            </div>
          </div>
          <div className='flex shrink-0 items-center justify-end gap-2 md:min-w-[8em]'>
            <div className='w-1/3 rounded-xl h-4 dark:bg-dark-monetr-background-subtle' />
          </div>
        </div>
      </li>
    );
  }

  if (isError) {
    return null;
  }

  const redirectUrl: string = `/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`;

  if (props.disableNavigate) {
    return (
      <Item>
        <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
        <ItemContent orientation='column' gap='none' flex='shrink' justify='start' align='default' shrink='default'>
          <Typography component='p' color='emphasis' size='md' weight='semibold' ellipsis>
            {transaction.getName()}
          </Typography>
          <Typography component='p' size='sm' weight='medium' ellipsis>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
          </Typography>
        </ItemContent>
        <ItemContent align='center' justify='end' flex='grow' shrink='none' width='fit'>
          <TransactionAmount transaction={transaction} />
        </ItemContent>
      </Item>
    );
  }

  return (
    <Item>
      <Link className={flexVariants({ orientation: 'row' })} to={redirectUrl}>
        <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
        <ItemContent orientation='column' gap='none' flex='shrink' justify='start' align='default' shrink='default'>
          <Typography component='p' color='emphasis' size='md' weight='semibold' ellipsis>
            {transaction.getName()}
          </Typography>
          <Typography component='p' size='sm' weight='medium' ellipsis>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
          </Typography>
        </ItemContent>
        <ItemContent align='center' justify='end' flex='grow' shrink='none' width='fit'>
          <TransactionAmount transaction={transaction} />
          <ArrowLink to={redirectUrl} />
        </ItemContent>
      </Link>
    </Item>
  );
}
