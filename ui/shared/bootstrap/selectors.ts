import { Plan } from 'shared/bootstrap/state';
import { AppState } from 'store';

export const getIsBootstrapped = (state: AppState): boolean => state.bootstrap.isReady;

export const getAPIUrl = (state: AppState): string => state.bootstrap.apiUrl;

export const getSignUpAllowed = (state: AppState): boolean => state.bootstrap.allowSignUp;

export const getShouldVerifyRegister = (state: AppState): boolean => state.bootstrap.ReCAPTCHAKey && state.bootstrap.verifyRegister;

export const getShouldVerifyLogin = (state: AppState): boolean => state.bootstrap.ReCAPTCHAKey && state.bootstrap.verifyLogin;

export const getShouldVerifyForgotPassword = (state: AppState): boolean => state.bootstrap.ReCAPTCHAKey && state.bootstrap.verifyForgotPassword;

export const getReCAPTCHAKey = (state: AppState): string | null => state.bootstrap.ReCAPTCHAKey;

export const getRequireLegalName = (state: AppState): boolean => state.bootstrap.requireLegalName;

export const getRequirePhoneNumber = (state: AppState): boolean => state.bootstrap.requirePhoneNumber;

export const getRequireBetaCode = (state: AppState): boolean => state.bootstrap.requireBetaCode;

export const getInitialPlan = (state: AppState): Plan | null => state.bootstrap.initialPlan || null;

export const getBillingEnabled = (state: AppState): boolean => state.bootstrap.billingEnabled;

export const getRelease = (state: AppState): string => state.bootstrap.release;

export const getRevision = (state: AppState): string => state.bootstrap.revision;

export const getAllowForgotPassword = (state: AppState) => state.bootstrap.allowForgotPassword;
