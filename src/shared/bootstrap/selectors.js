export const getIsBootstrapped = state => state.bootstrap.isReady;

export const getAPIUrl = state => state.bootstrap.apiUrl;

export const getSignUpAllowed = state => state.bootstrap.allowSignUp;

export const getShouldVerifyRegister = state => state.bootstrap.ReCAPTCHAKey && state.bootstrap.verifyRegister;

export const getShouldVerifyLogin = state => state.bootstrap.ReCAPTCHAKey && state.bootstrap.verifyLogin;

export const getReCAPTCHAKey = state => state.bootstrap.ReCAPTCHAKey;

export const getRequireLegalName = state => state.bootstrap.requireLegalName;

export const getRequirePhoneNumber = state => state.bootstrap.requirePhoneNumber;
