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
}) {

}
