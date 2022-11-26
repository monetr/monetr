import { useQuery, UseQueryResult } from 'react-query';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { SpendingType } from 'models/Spending';
import request from 'util/request';

interface SpendingBareMinimum {
  bankAccountId: number;
  nextRecurrence: moment.Moment;
  spendingType: SpendingType;
  fundingScheduleId: number;
  targetAmount: number;
  recurrenceRule: string | null,
}

interface SpendingForecast {
  estimatedCost: number;
}

export function useSpendingForecast(): (spending: SpendingBareMinimum) => Promise<SpendingForecast> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return async function (spending: SpendingBareMinimum): Promise<SpendingForecast> {
    return request()
      .post<SpendingForecast>(`/bank_accounts/${ selectedBankAccountId }/forecast/spending`, spending)
      .then(result => result.data);
  };
}

interface NextFundingResponse {
  nextContribution: number;
}

export type NextFundingResult =
  { result: number | null }
  & UseQueryResult<Partial<NextFundingResponse>>;

export function useNextFundingForecast(fundingScheduleId: number): NextFundingResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Partial<NextFundingResponse>>(
    [
      `/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`,
      {
        fundingScheduleId,
      },
    ],
    {
      enabled: !!selectedBankAccountId,
    }
  );

  return {
    ...result,
    result: result?.data?.nextContribution,
  };
}
