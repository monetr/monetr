import { Fragment } from 'react';
import { Link } from 'react-router-dom';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import { flexVariants } from '@monetr/interface/components/Flex';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import Typography from '@monetr/interface/components/Typography';
import TransactionAmount from '@monetr/interface/components/transactions/TransactionAmount';
import TransactionMerchantIcon from '@monetr/interface/components/transactions/TransactionMerchantIcon';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useSpendingTransactions from '@monetr/interface/hooks/useSpendingTransactions';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type Spending from '@monetr/interface/models/Spending';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

export interface ExpenseTransactionListProps {
  spending: Spending;
}

export default function ExpenseTransactionList(props: ExpenseTransactionListProps): React.JSX.Element {
  const { data: transactions } = useSpendingTransactions(props.spending.spendingId);
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocale();
  return (
    <Fragment>
      <Typography color='emphasis' size='xl' weight='medium' className='mb-4'>
        Transactions
      </Typography>
      <ol
        className={flexVariants({
          shrink: 'default',
          orientation: 'column',
        })}
      >
        {transactions?.map(transaction => (
          <Item key={transaction.transactionId}>
            <Link
              className={flexVariants({ orientation: 'row', align: 'center' })}
              to={`/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`}
            >
              <TransactionMerchantIcon name={transaction.getName()} pending={transaction.isPending} />
              <ItemContent
                orientation='column'
                gap='none'
                flex='shrink'
                justify='start'
                align='default'
                shrink='default'
              >
                <Typography component='p' color='emphasis' size='md' weight='semibold' ellipsis>
                  {transaction.getName()}
                </Typography>
                <Typography component='p' size='sm' weight='medium' ellipsis>
                  {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
                </Typography>
              </ItemContent>
              <ItemContent align='center' justify='end' flex='grow' shrink='none' width='fit'>
                <TransactionAmount transaction={transaction} />
                <ArrowLink
                  to={`/bank/${transaction.bankAccountId}/transactions/${transaction.transactionId}/details`}
                />
              </ItemContent>
            </Link>
          </Item>
        ))}
      </ol>
    </Fragment>
  );
}
