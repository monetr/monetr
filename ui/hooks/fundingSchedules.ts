import { useMutation, useQuery, useQueryClient, UseQueryResult } from 'react-query';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import FundingSchedule from 'models/FundingSchedule';
import request from 'util/request';

export type FundingSchedulesResult =
  { result: Map<number, FundingSchedule> }
  & UseQueryResult<Array<Partial<FundingSchedule>>>;

export function useFundingSchedulesSink(): FundingSchedulesResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Array<Partial<FundingSchedule>>>(
    `/bank_accounts/${ selectedBankAccountId }/funding_schedules`,
    {
      enabled: !!selectedBankAccountId,
    },
  );
  return {
    ...result,
    result: new Map(result?.data?.map(item => {
      const fundingSchedule = new FundingSchedule(item);
      return [fundingSchedule.fundingScheduleId, fundingSchedule];
    })),
  };
}

export function useFundingSchedules(): Map<number, FundingSchedule> {
  const { result } = useFundingSchedulesSink();
  return result;
}

export function useFundingSchedule(fundingScheduleId: number | null): FundingSchedule | null {
  const { result } = useFundingSchedulesSink();
  if (!fundingScheduleId) return null;

  return result.get(fundingScheduleId) || null;
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
          `/bank_accounts/${ newFundingSchedule.bankAccountId }/funding_schedules`,
          (previous: Array<Partial<FundingSchedule>>) => previous.concat(newFundingSchedule),
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
          `/bank_accounts/${ updatedFundingSchedule.bankAccountId }/funding_schedules`,
          (previous: Array<Partial<FundingSchedule>>) => previous.map(item =>
            item.fundingScheduleId === updatedFundingSchedule.fundingScheduleId ? updatedFundingSchedule : item
          ),
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
          `/bank_accounts/${ removed.bankAccountId }/funding_schedules`,
          (previous: Array<Partial<FundingSchedule>>) => previous
            .filter(item => item.fundingScheduleId !== removed.fundingScheduleId),
        ),
      ]),
    },
  );

  return async function (fundingSchedule: FundingSchedule): Promise<void> {
    return mutation.mutateAsync(fundingSchedule).then(() => { return; });
  };
}
