import { Link } from 'react-router-dom';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import { useTransaction } from '@monetr/interface/hooks/useTransaction';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from '../Item.module.scss';
import useTimezone from '@monetr/interface/hooks/useTimezone';

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

  return (
    <li className={styles.itemRoot}>
      <Link to={redirectUrl} className={styles.itemLink}>
        <Flex orientation='row' align='center' gap='lg' flex='shrink'>
          <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
          <Flex orientation='column' gap='none' flex='shrink'>
            <Typography color='emphasis' size='md' weight='semibold' ellipsis>
              {transaction.getName()}
            </Typography>
            <Typography size='sm' weight='medium' ellipsis>
              {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
            </Typography>
          </Flex>
        </Flex>
        <Flex align='center' justify='end' flex='grow' shrink='none' width='fit'>
          <TransactionAmount transaction={transaction} />
          {!props.disableNavigate && <ArrowLink to={redirectUrl} />}
        </Flex>
      </Link>
    </li>
  );
}
