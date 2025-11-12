import { Fragment } from 'react';
import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';

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
      <Typography className='mb-4' color='emphasis' size='xl' weight='medium'>
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
                align='default'
                flex='shrink'
                gap='none'
                justify='start'
                orientation='column'
                shrink='default'
              >
                <Typography color='emphasis' component='p' ellipsis size='md' weight='semibold'>
                  {transaction.getName()}
                </Typography>
                <Typography component='p' ellipsis size='sm' weight='medium'>
                  {formatDate(transaction.date, inTimezone, locale, DateLength.Long)}
                </Typography>
              </ItemContent>
              <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
                <TransactionAmount transaction={transaction} />
                <Typography>
                  <ChevronRight />
                </Typography>
              </ItemContent>
            </Link>
          </Item>
        ))}
      </ol>
    </Fragment>
  );
}
