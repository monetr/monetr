import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export class Forecast {
  startingTime: Date;
  endingTime: Date;
  startingBalance: number;
  endingBalance: number;
  events: Array<ForecastEvent>;

  constructor(data: WithJsonValues<Forecast>) {
    this.startingTime = parseDate(data.startingTime);
    this.endingTime = parseDate(data.endingTime);
    this.startingBalance = data.startingBalance;
    this.endingBalance = data.endingBalance;
    this.events = (data.events ?? []).map(item => new ForecastEvent(item));
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

  constructor(data: WithJsonValues<ForecastEvent>) {
    this.balance = data.balance;
    this.contribution = data.contribution;
    this.date = parseDate(data.date);
    this.delta = data.delta;
    this.funding = (data.funding ?? []).map(item => new FundingEvent(item));
    this.spending = (data.spending ?? []).map(item => new SpendingEvent(item));
    this.transaction = data.transaction;
  }
}

export class SpendingEvent {
  contributionAmount: number;
  date: Date;
  funding: Array<FundingEvent>;
  rollingAllocation: number;
  spendingId: string;
  transactionAmount: number;

  constructor(data: WithJsonValues<SpendingEvent>) {
    this.contributionAmount = data.contributionAmount;
    this.date = parseDate(data.date);
    this.funding = (data.funding ?? []).map(item => new FundingEvent(item));
    this.rollingAllocation = data.rollingAllocation;
    this.spendingId = data.spendingId;
    this.transactionAmount = data.transactionAmount;
  }
}

export class FundingEvent {
  date: Date;
  fundingScheduleId: string;
  originalDate: Date;
  weekendAvoided: boolean;

  constructor(data: WithJsonValues<FundingEvent>) {
    this.date = parseDate(data.date);
    this.fundingScheduleId = data.fundingScheduleId;
    this.originalDate = parseDate(data.originalDate);
    this.weekendAvoided = data.weekendAvoided;
  }
}

export function useForecast(): UseQueryResult<Forecast, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<WithJsonValues<Forecast>, unknown, Forecast>({
    queryKey: [`/api/bank_accounts/${selectedBankAccountId}/forecast`],
    enabled: Boolean(selectedBankAccountId),
    select: data => new Forecast(data),
  });
}
