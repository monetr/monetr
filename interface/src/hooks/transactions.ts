import {
  InfiniteData,
  useInfiniteQuery,
  UseInfiniteQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
  UseQueryResult,
} from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Balance from '@monetr/interface/models/Balance';
import Spending from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import TransactionCluster from '@monetr/interface/models/TransactionCluster';
import request from '@monetr/interface/util/request';

export type TransactionsResult = {
  result: Array<Transaction>;
  hasNextPage: boolean;
} & UseInfiniteQueryResult<Array<Partial<Transaction>>>;

export function useTransactions(): TransactionsResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useInfiniteQuery<Array<Partial<Transaction>>>(
    [`/bank_accounts/${ selectedBankAccountId }/transactions`],
    {
      getNextPageParam: (_, pages) => pages.length * 25,
      enabled: !!selectedBankAccountId,
    },
  );
  return {
    ...result,
    hasNextPage: !result?.data?.pages.some(page => page.length < 25),
    // Take all the pages and build an array. Make sure we actually return an array here even if it's empty.
    result: result?.data?.pages.flatMap(x => x).map(item => new Transaction(item)) || [],
  };
}

export function useTransaction(transactionId: string | null): UseQueryResult<Transaction> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Transaction>, unknown, Transaction>(
    [`/bank_accounts/${ selectedBankAccountId }/transactions/${ transactionId }`],
    {
      enabled: !!selectedBankAccountId && !!transactionId,
      select: data => new Transaction(data),
    },
  );
}

export function useSimilarTransactions(transaction: Transaction | null): UseQueryResult<TransactionCluster> {
  return useQuery<Partial<TransactionCluster>, unknown, TransactionCluster>(
    [`/bank_accounts/${ transaction?.bankAccountId }/transactions/${ transaction?.transactionId }/similar`],
    {
      enabled: Boolean(transaction),
      select: data => new TransactionCluster(data),
      retry: false,
    },
  );
}

export function useSpendingTransactions(spending?: Spending): UseQueryResult<Array<Transaction>> {
  return useQuery<Array<Partial<Transaction>>, unknown, Array<Transaction>>(
    [`/bank_accounts/${ spending?.bankAccountId }/spending/${ spending?.spendingId }/transactions`],
    {
      enabled: Boolean(spending),
      select: data => data.map(item => new Transaction(item)),
    },
  );
}

export interface TransactionUpdateResponse {
  transaction: Partial<Transaction>;
  spending: Array<Partial<Spending>>;
  balance: Partial<Balance>;
}

export function useUpdateTransaction(): (_transaction: Transaction) => Promise<TransactionUpdateResponse> {
  const queryClient = useQueryClient();

  async function updateTransaction(transaction: Transaction): Promise<TransactionUpdateResponse> {
    return request()
      .put<TransactionUpdateResponse>(
        `/bank_accounts/${ transaction.bankAccountId }/transactions/${ transaction.transactionId }`,
        transaction,
      )
      .then(result => result.data);
  }

  const { mutateAsync } = useMutation(
    updateTransaction,
    {
      onSuccess: (response: TransactionUpdateResponse) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/transactions`],
          (previous: InfiniteData<Array<Transaction>>) => ({
            ...previous,
            pages: previous.pages.map(page =>
              page.map(item =>
                item.transactionId === response.transaction.transactionId ? response.transaction : item
              )
            ),
          }),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/transactions/${response.transaction.transactionId}`],
          response.transaction,
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => (response.spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        (response.spending || []).map(spending =>
          queryClient.setQueriesData(
            [`/bank_accounts/${response.transaction.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          )),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/balances`],
          (previous: Partial<Balance>) => new Balance({
            ...previous,
            ...response.balance,
          }),
        ),
        queryClient.invalidateQueries([`/bank_accounts/${ response.transaction.bankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ response.transaction.bankAccountId }/forecast/next_funding`]),
      ]),
    }
  );

  return mutateAsync;
}

export interface CreateTransactionRequest {
  name: string;
  bankAccountId: string;
  amount: number;
  spendingId: string | null;
  date: Date;
  merchantName: string | null;
  isPending: boolean;
  adjustsBalance: boolean;
}

export interface CreateTransactionResponse {
  transaction: Partial<Transaction>;
  balance: Partial<Balance>;
  spending: Partial<Spending> | null;
}

export function useCreateTransaction(): (_: CreateTransactionRequest) => Promise<CreateTransactionResponse> {
  const queryClient = useQueryClient();

  async function createTransaction(transaction: CreateTransactionRequest): Promise<CreateTransactionResponse> {
    return request()
      .post<CreateTransactionResponse>(
        `/bank_accounts/${ transaction.bankAccountId }/transactions`,
        transaction,
      )
      .then(result => result.data);
  }

  const { mutateAsync } = useMutation(
    createTransaction,
    {
      onSuccess: (response: CreateTransactionResponse) => Promise.all([
        queryClient.invalidateQueries([`/bank_accounts/${ response.transaction.bankAccountId }/transactions`]),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => response?.spending?.spendingId == item.spendingId ? response?.spending : item),
        ),
        response.spending != null && queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/spending/${response.spending.spendingId}`],
          response.spending,
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/balances`],
          (previous: Partial<Balance>) => new Balance({
            ...previous,
            ...response.balance,
          }),
        ),
      ]),
    }
  );

  return mutateAsync;
}

export interface RemoveTransactionRequest {
  transaction: Transaction;
  adjustsBalance: boolean;
  softDelete: boolean;
}

interface RemoveTransactionResponse {
  balance: Partial<Balance>;
  spending: Partial<Spending> | null;
}

export function useRemoveTransaction(): (_: RemoveTransactionRequest) => Promise<unknown> {
  const queryClient = useQueryClient();

  // TODO It would be better to make this a mutator, however; since the response does not include the transaction that
  // was removed it is tricky to do in place mutations. At the moment this implementation instead just invalidates the
  // related queries that might be affected by the removal.
  async function removeTransaction(removal: RemoveTransactionRequest): Promise<unknown> {
    const { transaction, softDelete, adjustsBalance } = removal;
    const path = `/bank_accounts/${transaction.bankAccountId}/transactions/${transaction.transactionId}`;
    const params = new URLSearchParams();
    { // Build the query parameters for the request.
      params.set('adjusts_balance', String(adjustsBalance));
      params.set('soft', String(softDelete));
    }
    // Send the delete request to the server and handle any changes returned.
    return await request()
      .delete(`${path}?${params.toString()}`)
      .then(result => result.data)
      .then(async (_: RemoveTransactionResponse) => await Promise.all([
        // TODO Instead of just invalidating all of these, it would be more efficient to move them into a mutator so we
        // can update their data in place.
        queryClient.invalidateQueries([`/bank_accounts/${ removal.transaction.bankAccountId }/transactions`]),
        queryClient.invalidateQueries([`/bank_accounts/${ removal.transaction.bankAccountId }/spending`]),
        queryClient.invalidateQueries([`/bank_accounts/${ removal.transaction.bankAccountId }/balances`]),
      ]));
  }

  return removeTransaction;
}
