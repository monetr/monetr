import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import parseDate from '@monetr/interface/util/parseDate';

export class Forecast {
  startingTime: Date;
  endingTime: Date;
  startingBalance: number;
  endingBalance: number;
  events: Array<ForecastEvent>;

  constructor(data?: Partial<Forecast>) {
    if (data)
      Object.assign(this, {
        ...data,
        startingTime: parseDate(data.startingTime),
        endingTime: parseDate(data.endingTime),
        events: (data?.events || []).map(item => new ForecastEvent(item)),
      });
  }
}

export class ForecastEvent {
  balance: number;
  contribution: number;
  date: Date;
  delta: number;
  funding: Array<FundingEvent>;
  spending: Array<SpendingEvent>;
  transaction: number;

  constructor(data?: Partial<ForecastEvent>) {
    if (data)
      Object.assign(this, {
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
    if (data)
      Object.assign(this, {
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
    if (data)
      Object.assign(this, {
        ...data,
        date: parseDate(data.date),
        originalDate: parseDate(data.originalDate),
      });
  }
}

export function useForecast(): UseQueryResult<Forecast, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Forecast>, unknown, Forecast>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/forecast`],
    enabled: Boolean(selectedBankAccountId),
    select: data => new Forecast(data),
  });
}
