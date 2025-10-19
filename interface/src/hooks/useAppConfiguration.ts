import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import parseDate from '@monetr/interface/util/parseDate';

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
  plaidEnabled: boolean;
  manualEnabled: true;
  uploadsEnabled: boolean;
  release: string | null;
  revision: string;
  buildType: string;
  buildTime: Date | null;

  constructor(data?: Partial<AppConfiguration>) {
    if (data) {
      Object.assign(this, {
        ...data,
        buildTime: parseDate(data?.buildTime),
      });
    }
  }
}

export function useAppConfiguration(): UseQueryResult<AppConfiguration, unknown> {
  return useQuery<Partial<AppConfiguration>, unknown, AppConfiguration>({
    queryKey: ['/config'],
    staleTime: 60 * 1000, // One minute in milliseconds.
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
    select: data => new AppConfiguration(data),
  });
}

export function useAppConfigurationBAD(): AppConfiguration | null {
  const { result } = useAppConfiguration();
  return result;
}
