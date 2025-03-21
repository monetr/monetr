import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import { SpendingType } from '@monetr/interface/models/Spending';
import parseDate from '@monetr/interface/util/parseDate';
import request from '@monetr/interface/util/request';

interface SpendingBareMinimum {
  bankAccountId: string;
  nextRecurrence: Date;
  spendingType: SpendingType;
  fundingScheduleId: string;
  targetAmount: number;
  recurrenceRule: string | null,
}

interface SpendingForecast {
  estimatedCost: number;
}

export function useSpendingForecast(): (spending: SpendingBareMinimum) => Promise<SpendingForecast> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return async function (spending: SpendingBareMinimum): Promise<SpendingForecast> {
    return await request()
      .post<SpendingForecast>(`/bank_accounts/${ selectedBankAccountId }/forecast/spending`, spending)
      .then(result => result.data);
  };
}

export function useNextFundingForecast(fundingScheduleId: string): UseQueryResult<number> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<{ nextContribution: number }>, unknown, number>(
    [
      `/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`,
      {
        fundingScheduleId,
      },
    ],
    {
      enabled: !!selectedBankAccountId,
      select: data => data.nextContribution,
    }
  );
}

export class Forecast {
  startingTime: Date;
  endingTime: Date;
  startingBalance: number;
  endingBalance: number;
  events: Array<Event>;

  constructor(data?: Partial<Forecast>) {
    if (data) Object.assign(this, {
      ...data,
      startingTime: parseDate(data.startingTime),
      endingTime: parseDate(data.endingTime),
      events: (data?.events || []).map(item => new Event(item)),
    });
  }
}

export class Event {
  balance: number;
  contribution: number;
  date: Date;
  delta: number;
  funding: Array<FundingEvent>;
  spending: Array<SpendingEvent>;
  transaction: number;

  constructor(data?: Partial<Event>) {
    if (data) Object.assign(this, {
      ...data,
      date: parseDate(data.date),
      funding: (data?.funding || []).map(item => new FundingEvent(item)),
      spending: (data?.spending || []).map(item => new SpendingEvent(item)),
    });
  }
}

export class SpendingEvent {
  contributionAmount: number;
  date: Date;
  funding: Array<FundingEvent>;
  rollingAllocation: number;
  spendingId: string;
  transactionAmount: number;

  constructor(data?: Partial<SpendingEvent>) {
    if (data) Object.assign(this, {
      ...data,
      date: parseDate(data.date),
      funding: (data?.funding || []).map(item => new FundingEvent(item)),
    });
  }
}

export class FundingEvent {
  date: Date;
  fundingScheduleId: string;
  originalDate: Date;
  weekendAvoided: boolean;

  constructor(data?: Partial<FundingEvent>) {
    if (data) Object.assign(this, {
      ...data,
      date: parseDate(data.date),
      originalDate: parseDate(data.originalDate),
    });
  }
}

export type ForecastResult =
  { result: Forecast | null }
  & UseQueryResult<Partial<Forecast>>;

export function useForecast(): ForecastResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Partial<Forecast>>(
    [`/bank_accounts/${ selectedBankAccountId }/forecast`],
    {
      // TODO long cache time for forecast endpoints.
      enabled: !!selectedBankAccountId,
    }
  );

  return {
    ...result,
    result: result?.data ? new Forecast(result.data) : null,
  };
}
