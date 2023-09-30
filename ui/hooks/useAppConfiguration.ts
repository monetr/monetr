import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { parseJSON } from 'date-fns';

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
  buildTime: Date | null;

  constructor(data?: Partial<AppConfiguration>) {
    if (data) Object.assign(this, {
      ...data,
      buildTime: data.buildTime && parseJSON(data.buildTime),
    });
  }
}

export type AppConfigurationResult =
  { result: AppConfiguration | null }
  & UseQueryResult<Partial<AppConfiguration>, unknown>;

export function useAppConfigurationSink(): AppConfigurationResult {
  const result = useQuery<Partial<AppConfiguration>>(['/config'], {
    staleTime: 60 * 1000, // One minute in milliseconds.
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
  });
  return {
    result: (result?.data && new AppConfiguration(result.data)) || null,
    ...result,
  };
}

export function useAppConfiguration(): AppConfiguration | null {
  const { result } = useAppConfigurationSink();
  return result;
}
