import {Record} from 'immutable';

export default class BootstrapState extends Record({
  apiUrl: '',
  isReady: false,
  isBootstrapping: true,
  verifyLogin: false,
  verifyRegister: false,
  ReCAPTCHAKey: null,
  allowSignUp: false,
  allowForgotPassword: false,
}) {

}
