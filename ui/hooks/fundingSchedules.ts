/* eslint-disable max-len */
import { useMutation, useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import FundingSchedule from 'models/FundingSchedule';
import request from 'util/request';

export type FundingSchedulesResult = UseQueryResult<Array<FundingSchedule>, unknown>;

export function useFundingSchedulesSink(): FundingSchedulesResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<FundingSchedule>>, unknown, Array<FundingSchedule>>(
    [`/bank_accounts/${ selectedBankAccountId }/funding_schedules`],
    {
      enabled: !!selectedBankAccountId,
      select: data => data?.map(item => new FundingSchedule(item)),
    },
  );
}

export type FundingScheduleResult = UseQueryResult<FundingSchedule | undefined, unknown>;

export function useFundingSchedule(fundingScheduleId: number | null): FundingScheduleResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<FundingSchedule>, unknown, FundingSchedule | null>(
    [`/bank_accounts/${ selectedBankAccountId }/funding_schedules/${ fundingScheduleId }`],
    {
      enabled: !!selectedBankAccountId && !!fundingScheduleId,
      select: data => data?.fundingScheduleId ? new FundingSchedule(data) : null,
    },
  );
}

export function useCreateFundingSchedule(): (_spending: FundingSchedule) => Promise<void> {
  const queryClient = useQueryClient();

  async function createFundingSchedule(newItem: FundingSchedule): Promise<FundingSchedule> {
    return request()
      .post<Partial<FundingSchedule>>(`/bank_accounts/${ newItem.bankAccountId }/funding_schedules`, newItem)
      .then(result => new FundingSchedule(result?.data));
  }

  const { mutate } = useMutation(
    createFundingSchedule,
    {
      onSuccess: (newFundingSchedule: FundingSchedule) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${ newFundingSchedule.bankAccountId }/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous.concat(newFundingSchedule),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${ newFundingSchedule.bankAccountId }/funding_schedules/${ newFundingSchedule.fundingScheduleId }`],
          newFundingSchedule,
        ),
      ]),
    },
  );

  return async (spending: FundingSchedule): Promise<void> => {
    return mutate(spending);
  };
}

export function useUpdateFundingSchedule(): (_fundingSchedule: FundingSchedule) => Promise<FundingSchedule> {
  const queryClient = useQueryClient();

  async function updateFundingSchedule(fundingSchedule: FundingSchedule): Promise<FundingSchedule> {
    return request()
      .put<Partial<FundingSchedule>>(
        `/bank_accounts/${ fundingSchedule.bankAccountId }/funding_schedules/${ fundingSchedule.fundingScheduleId }`,
        fundingSchedule,
      )
      .then(result => new FundingSchedule(result?.data));
  }

  const mutation = useMutation(
    updateFundingSchedule,
    {
      onSuccess: (updatedFundingSchedule: FundingSchedule) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${ updatedFundingSchedule.bankAccountId }/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous.map(item =>
            item.fundingScheduleId === updatedFundingSchedule.fundingScheduleId ? updatedFundingSchedule : item
          ),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${ updatedFundingSchedule.bankAccountId }/funding_schedules/${ updatedFundingSchedule.fundingScheduleId}`],
          updatedFundingSchedule,
        ),
      ]),
    },
  );

  return mutation.mutateAsync;
}

export function useRemoveFundingSchedule(): (_fundingSchedule: FundingSchedule) => Promise<void> {
  const queryClient = useQueryClient();

  async function removeFundingSchedule(fundingSchedule: FundingSchedule): Promise<FundingSchedule> {
    return request()
      .delete(
        `/bank_accounts/${ fundingSchedule.bankAccountId }/funding_schedules/${ fundingSchedule.fundingScheduleId }`,
      )
      .then(() => fundingSchedule);
  }

  const mutation = useMutation(
    removeFundingSchedule,
    {
      onSuccess: (removed: FundingSchedule) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${ removed.bankAccountId }/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous
            .filter(item => item.fundingScheduleId !== removed.fundingScheduleId),
        ),
        queryClient.removeQueries(
          [`/bank_accounts/${ removed.bankAccountId }/funding_schedules/${ removed.fundingScheduleId }`]
        ),
      ]),
    },
  );

  return async function (fundingSchedule: FundingSchedule): Promise<void> {
    return mutation.mutateAsync(fundingSchedule).then(() => { return; });
  };
}
