import { useQuery } from 'react-query';

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

export function useAppConfiguration(): AppConfiguration {
  const { data } = useQuery<Partial<AppConfiguration>>('/api/config');
  return new AppConfiguration(data);
}
