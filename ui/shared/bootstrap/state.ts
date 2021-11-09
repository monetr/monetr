export class Plan {
  price: number;
  freeTrialDays: number;
}

export default class BootstrapState {
  readonly apiUrl: string;
  readonly isReady: boolean;
  readonly isBootstrapping: boolean;
  readonly verifyLogin: boolean;
  readonly verifyRegister: boolean;
  readonly requireLegalName: boolean;
  readonly requirePhoneNumber: boolean;
  readonly ReCAPTCHAKey: string | null;
  readonly allowSignUp: boolean;
  readonly allowForgotPassword: boolean;
  readonly requireBetaCode: boolean;
  readonly stripePublicKey: string | null;
  readonly initialPlan: Plan | null;
  readonly billingEnabled: boolean;
  readonly release: string;
  readonly revision: string;

  constructor(data?: Partial<BootstrapState>) {
    if (data) {
      Object.assign(this, data);
    }

    this.isBootstrapping = data?.isBootstrapping || true;
  }
}
