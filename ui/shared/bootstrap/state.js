import { Record } from 'immutable';

export default class BootstrapState extends Record({
  apiUrl: '',
  isReady: false,
  isBootstrapping: true,
  verifyLogin: false,
  verifyRegister: false,
  requireLegalName: false,
  requirePhoneNumber: false,
  ReCAPTCHAKey: null,
  allowSignUp: false,
  allowForgotPassword: false,
  requireBetaCode: false,
  stripePublicKey: '',
  initialPlan: {
    price: 0,
    freeTrialDays: 0,
  },
  billingEnabled: false,
  release: '',
  revision: '',
}) {

}
