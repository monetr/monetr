import { useQuery, UseQueryResult } from 'react-query';

import moment from 'moment';
import { parseToMomentMaybe } from 'util/parseToMoment';

export class AppConfiguration {
  requireLegalName: boolean;
  requirePhoneNumber: boolean;
  verifyLogin: boolean;
  verifyRegister: boolean;
  verifyEmailAddress: boolean;
  verifyForgotPassword: boolean;
  ReCAPTCHAKey: string | null;
  allowSignUp: boolean;
  allowForgotPassword: boolean;
  longPollPlaidSetup: boolean;
  requireBetaCode: boolean;
  initialPlan: {
    price: number;
    freeTrialDays: number;
  } | null;
  billingEnabled: boolean;
  iconsEnabled: boolean;
  release: string | null;
  revision: string;
  buildType: string;
  buildTime: moment.Moment | null;

  constructor(data?: Partial<AppConfiguration>) {
    if (data) Object.assign(this, {
      ...data,
      buildTime: parseToMomentMaybe(data.buildTime),
    });
  }
}

export interface AppConfigurationWrapper {
  result: AppConfiguration;
}

export type AppConfigurationResult = AppConfigurationWrapper & UseQueryResult<Partial<AppConfiguration>, unknown>;

export function useAppConfigurationSink(): AppConfigurationResult {
  const result = useQuery<Partial<AppConfiguration>>('/api/config', {
    staleTime: 60 * 1000, // One minute in milliseconds.
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
  });
  return {
    result: new AppConfiguration(result.data),
    ...result,
  };
}

export function useAppConfiguration(): AppConfiguration {
  const { result } = useAppConfigurationSink();
  return result;
}
