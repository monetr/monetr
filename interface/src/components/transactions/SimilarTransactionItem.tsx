import { ChevronRight } from 'lucide-react';
import { Link } from 'wouter';

import Badge from '@monetr/interface/components/Badge';
import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type Transaction from '@monetr/interface/models/Transaction';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './SimilarTransactionItem.module.scss';

export interface SimilarTransactionItemProps {
  transaction: Transaction;
  /**
   * disableNavigate will remove the arrow link or the click-ability of the similar transaction item.
   */
  disableNavigate?: boolean;
  /**
   * current indicates that this item is the transaction currently being viewed, and will show a badge to call that out.
   */
  current?: boolean;
}

export default function SimilarTransactionItem(props: SimilarTransactionItemProps): React.JSX.Element | null {
  const { transaction } = props;
  const { inTimezone } = useTimezone();
  const { data: locale, isLoading: localeIsLoading } = useLocale();

  if (localeIsLoading) {
    return (
      <li className={styles.root}>
        <div className={styles.skeleton}>
          <div className={styles.leftSection}>
            <div className={styles.iconPlaceholder} />
            <div className={styles.textColumn}>
              <div className={styles.linePrimary} />
              <div className={styles.lineSecondary} />
            </div>
          </div>
          <div className={styles.amountSection}>
            <div className={styles.amountPlaceholder} />
          </div>
        </div>
      </li>
    );
  }

  if (!locale) {
    return null;
  }

  const redirectUrl: string = `/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`;

  if (props.disableNavigate) {
    return (
      <Item>
        <TransactionMerchantIcon
          name={transaction.getName()}
          pending={transaction.isPending}
          transactionRecurringId={transaction.transactionRecurringId}
        />
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <Typography color='emphasis' component='p' ellipsis size='md' weight='semibold'>
            {transaction.getName()}
          </Typography>
          <Typography component='p' ellipsis size='sm' weight='medium'>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
          </Typography>
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          {props.current && (
            <Badge size='xs' variant='info' wrapping='nowrap'>
              This One
            </Badge>
          )}
          <TransactionAmount transaction={transaction} />
        </ItemContent>
      </Item>
    );
  }

  return (
    <Item>
      <Link className={flexVariants({ orientation: 'row', align: 'center' })} to={redirectUrl}>
        <TransactionMerchantIcon
          name={transaction.getName()}
          pending={transaction.isPending}
          transactionRecurringId={transaction.transactionRecurringId}
        />
        <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
          <Typography color='emphasis' component='p' ellipsis size='md' weight='semibold'>
            {transaction.getName()}
          </Typography>
          <Typography component='p' ellipsis size='sm' weight='medium'>
            {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
          </Typography>
        </ItemContent>
        <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
          {props.current && (
            <Badge size='xs' variant='info' wrapping='nowrap'>
              This One
            </Badge>
          )}
          <TransactionAmount transaction={transaction} />
          <Typography>
            <ChevronRight />
          </Typography>
        </ItemContent>
      </Link>
    </Item>
  );
}
