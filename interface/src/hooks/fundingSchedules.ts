/* eslint-disable max-len */
import { useMutation, useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';
import { AxiosResponse } from 'axios';
import { format, isBefore } from 'date-fns';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export function useFundingSchedulesSink(): UseQueryResult<Array<FundingSchedule>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<FundingSchedule>>, unknown, Array<FundingSchedule>>(
    [`/bank_accounts/${selectedBankAccountId}/funding_schedules`],
    {
      enabled: !!selectedBankAccountId,
      select: data => data.map(item => new FundingSchedule(item)),
    },
  );
}

/**
 *  useNextFundingDate will return a M/DD formatted string showing when the next funding schedule will recur. This is
 *  just the earliest funding shedule among all the funding schedules for the current bank account.
 */
export function useNextFundingDate(): string | null {
  const { data: funding } = useFundingSchedulesSink();
  const date = funding
    ?.sort((a, b) => isBefore(a.nextRecurrence, b.nextRecurrence) ? 1 : -1)
    .pop();

  if (date) {
    return format(date.nextRecurrence, 'M/dd');
  }

  return null;
}

export function useFundingSchedule(fundingScheduleId: string | null): UseQueryResult<FundingSchedule | undefined, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<FundingSchedule>, unknown, FundingSchedule | null>(
    [`/bank_accounts/${selectedBankAccountId}/funding_schedules/${fundingScheduleId}`],
    {
      enabled: !!selectedBankAccountId && !!fundingScheduleId,
      select: data => data?.fundingScheduleId ? new FundingSchedule(data) : null,
    },
  );
}

export type CreateFundingScheduleRequest = Pick<FundingSchedule,
  'bankAccountId' |
  'name' |
  'description' |
  'ruleset' |
  'nextRecurrence' |
  'excludeWeekends' |
  'estimatedDeposit'
>;

export function useCreateFundingSchedule(): (_funding: CreateFundingScheduleRequest) => Promise<FundingSchedule> {
  const queryClient = useQueryClient();

  async function createFundingSchedule({
    bankAccountId,
    ...fundingSchedule
  }: CreateFundingScheduleRequest): Promise<FundingSchedule> {
    return request()
      .post<Partial<FundingSchedule>>(`/bank_accounts/${bankAccountId}/funding_schedules`, fundingSchedule)
      .then(result => new FundingSchedule(result?.data));
  }

  const mutate = useMutation(
    createFundingSchedule,
    {
      onSuccess: (newFundingSchedule: FundingSchedule) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${newFundingSchedule.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous.concat(newFundingSchedule),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${newFundingSchedule.bankAccountId}/funding_schedules/${newFundingSchedule.fundingScheduleId}`],
          newFundingSchedule,
        ),
      ]),
    },
  );

  return mutate.mutateAsync;
}

export type PatchFundingScheduleRequest =
  Pick<FundingSchedule, 'fundingScheduleId' | 'bankAccountId'> &
  Partial<Pick<FundingSchedule,
    'name' |
    'description' |
    'ruleset' |
    'nextRecurrence' |
    'excludeWeekends' |
    'estimatedDeposit'
  >>

export interface PatchFundingScheduleResponse {
  fundingSchedule: FundingSchedule;
  spending: Array<Spending>;
}

export function usePatchFundingSchedule(): (_patch: PatchFundingScheduleRequest) => Promise<PatchFundingScheduleResponse> {
  const queryClient = useQueryClient();

  async function patchFundingSchedule(
    { fundingScheduleId, bankAccountId, ...patch }: PatchFundingScheduleRequest,
  ): Promise<PatchFundingScheduleResponse> {
    return request()
      .patch<FundingSchedule, AxiosResponse<PatchFundingScheduleResponse>>(
        `/bank_accounts/${bankAccountId}/funding_schedules/${fundingScheduleId}`,
        patch,
      )
      .then(result => ({
        fundingSchedule: new FundingSchedule(result.data.fundingSchedule),
        spending: result.data.spending.map(spending => new Spending(spending)),
      }));
  }

  const mutation = useMutation(
    patchFundingSchedule,
    {
      onSuccess: (response: PatchFundingScheduleResponse) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous.map(item =>
            item.fundingScheduleId === response.fundingSchedule.fundingScheduleId ? response.fundingSchedule : item
          ),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/funding_schedules/${response.fundingSchedule.fundingScheduleId}`],
          response.fundingSchedule,
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => (response.spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        (response.spending || []).map(spending =>
          queryClient.setQueriesData(
            [`/bank_accounts/${response.fundingSchedule.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          )),
        queryClient.invalidateQueries([`/bank_accounts/${ response.fundingSchedule.bankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ response.fundingSchedule.bankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return mutation.mutateAsync;
}

export interface FundingScheduleUpdateResponse {
  fundingSchedule: Partial<FundingSchedule>;
  spending: Array<Partial<Spending>>;
}

export function useUpdateFundingSchedule(): (_fundingSchedule: FundingSchedule) => Promise<FundingScheduleUpdateResponse> {
  const queryClient = useQueryClient();

  async function updateFundingSchedule(fundingSchedule: FundingSchedule): Promise<FundingScheduleUpdateResponse> {
    return request()
      .put<FundingSchedule, AxiosResponse<FundingScheduleUpdateResponse>>(
        `/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules/${fundingSchedule.fundingScheduleId}`,
        fundingSchedule,
      )
      .then(result => result.data);
  }

  const mutation = useMutation(
    updateFundingSchedule,
    {
      onSuccess: (response: FundingScheduleUpdateResponse) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous.map(item =>
            item.fundingScheduleId === response.fundingSchedule.fundingScheduleId ? response.fundingSchedule : item
          ),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/funding_schedules/${response.fundingSchedule.fundingScheduleId}`],
          response.fundingSchedule,
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => (response.spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        (response.spending || []).map(spending =>
          queryClient.setQueriesData(
            [`/bank_accounts/${response.fundingSchedule.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          )),
        queryClient.invalidateQueries([`/bank_accounts/${ response.fundingSchedule.bankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ response.fundingSchedule.bankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return mutation.mutateAsync;
}

export function useRemoveFundingSchedule(): (_fundingSchedule: FundingSchedule) => Promise<FundingSchedule> {
  const queryClient = useQueryClient();

  async function removeFundingSchedule(fundingSchedule: FundingSchedule): Promise<FundingSchedule> {
    return request()
      .delete(
        `/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules/${fundingSchedule.fundingScheduleId}`,
      )
      .then(() => fundingSchedule);
  }

  const mutation = useMutation(
    removeFundingSchedule,
    {
      onSuccess: ({ bankAccountId, fundingScheduleId }: FundingSchedule) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous
            .filter(item => item.fundingScheduleId !== fundingScheduleId),
        ),
        queryClient.removeQueries(
          [`/bank_accounts/${bankAccountId}/funding_schedules/${fundingScheduleId}`]
        ),
      ]),
    },
  );

  return mutation.mutateAsync;
}
