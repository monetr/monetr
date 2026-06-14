import { useCallback } from 'react';
import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import type { WithJsonValues } from '@monetr/interface/util/json';
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
  lunchFlowEnabled: boolean;
  lunchFlowAllowedAPIURLs: Array<string>;
  manualEnabled: true;
  uploadsEnabled: boolean;
  release: string | null;
  revision: string;
  buildType: string;
  buildTime: Date | null;

  constructor(data: WithJsonValues<AppConfiguration>) {
    this.requireLegalName = data.requireLegalName;
    this.requirePhoneNumber = data.requirePhoneNumber;
    this.verifyLogin = data.verifyLogin;
    this.verifyRegister = data.verifyRegister;
    this.verifyEmailAddress = data.verifyEmailAddress;
    this.verifyForgotPassword = data.verifyForgotPassword;
    this.ReCAPTCHAKey = data.ReCAPTCHAKey ?? null;
    this.allowSignUp = data.allowSignUp;
    this.allowForgotPassword = data.allowForgotPassword;
    this.longPollPlaidSetup = data.longPollPlaidSetup;
    this.requireBetaCode = data.requireBetaCode;
    this.initialPlan = data.initialPlan ?? null;
    this.billingEnabled = data.billingEnabled;
    this.iconsEnabled = data.iconsEnabled;
    this.plaidEnabled = data.plaidEnabled;
    this.lunchFlowEnabled = data.lunchFlowEnabled;
    this.lunchFlowAllowedAPIURLs = data.lunchFlowAllowedAPIURLs;
    this.manualEnabled = data.manualEnabled;
    this.uploadsEnabled = data.uploadsEnabled;
    this.release = data.release ?? null;
    this.revision = data.revision;
    this.buildType = data.buildType;
    this.buildTime = parseDate(data.buildTime);
  }
}

export function useAppConfiguration(): UseQueryResult<AppConfiguration, unknown> {
  const select = useCallback((data: WithJsonValues<AppConfiguration>) => new AppConfiguration(data), []);
  return useQuery<WithJsonValues<AppConfiguration>, unknown, AppConfiguration>({
    queryKey: ['/api/config'],
    staleTime: 60 * 1000, // One minute in milliseconds.
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
    select,
    notifyOnChangeProps: ['data', 'isLoading', 'isError'],
  });
}
