import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { QueryMethod } from '@monetr/interface/components/MQueryClient';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';

export interface RecurringPriceChange {
  previousAmount: number;
  previousCount: number;
  currentAmount: number;
  currentCount: number;
  // changedAt is the date of the first transaction that came in at the new amount
  changedAt: Date;
}

export interface RecurringTransactionHistory {
  /**
   * transactions in the recurring transaction, newest first. its only the most recent 100 though so on long running
   * ones this can be less than seen
   */
  transactions: Array<Transaction>;
  /**
   * seen is how many transactions are in the recurring transaction total. comes from the amounts on the recurring
   * transaction itself so its always complete
   */
  seen: number;
  priceChange: RecurringPriceChange | null;
}

export function useRecurringTransactionHistory(
  recurring: TransactionRecurring | undefined,
): RecurringTransactionHistory & { isLoading: boolean } {
  const { data: transactions, isLoading } = useQuery<Array<WithJsonValues<Transaction>>, unknown, Array<Transaction>>({
    queryKey: [
      `/api/bank_accounts/${recurring?.bankAccountId}/transactions`,
      { transaction_recurring_id: recurring?.transactionRecurringId.toString(), limit: 100 },
    ],
    enabled: Boolean(recurring),
    meta: {
      method: QueryMethod.UseQuery,
    },
    select: data => data.map(item => new Transaction(item)),
  });

  return useMemo(() => {
    if (!recurring) {
      return { transactions: [], seen: 0, priceChange: null, isLoading };
    }

    return {
      ...getRecurringTransactionHistory(recurring, transactions ?? []),
      isLoading,
    };
  }, [recurring, transactions, isLoading]);
}

// getRecurringTransactionHistory expects the transactions to already be newest first, which is how the API gives them back
export function getRecurringTransactionHistory(
  recurring: TransactionRecurring,
  transactions: Array<Transaction>,
): RecurringTransactionHistory {
  const seen = Object.values(recurring.amounts).reduce((total, count) => total + count, 0);

  return {
    transactions,
    seen,
    priceChange: getRecurringPriceChange(recurring, transactions),
  };
}

function getRecurringPriceChange(
  recurring: TransactionRecurring,
  transactions: Array<Transaction>,
): RecurringPriceChange | null {
  // walk back from the newest transaction until we hit one with a different amount, the one right before that is where
  // the price changed
  const index = transactions.findIndex(item => item.amount !== transactions[0]?.amount);
  const current = transactions[index - 1];
  const previous = transactions[index];
  if (!current || !previous) {
    return null;
  }

  return {
    previousAmount: previous.amount,
    previousCount: recurring.amounts[previous.amount] ?? 0,
    currentAmount: current.amount,
    currentCount: recurring.amounts[current.amount] ?? 0,
    changedAt: current.date,
  };
}
