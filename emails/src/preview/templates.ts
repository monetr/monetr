import { VerifyEmailAddress } from '../emails/VerifyEmailAddress';
import { ForgotPassword } from '../emails/ForgotPassword';
import { PasswordChanged } from '../emails/PasswordChanged';
import { PlaidDisconnected } from '../emails/PlaidDisconnected';
import { TrialAboutToExpire } from '../emails/TrialAboutToExpire';

export interface TemplateEntry {
  name: string;
  component: React.ComponentType<any>;
  previewProps: Record<string, any>;
}

export const templates: TemplateEntry[] = [
  {
    name: 'VerifyEmailAddress',
    component: VerifyEmailAddress,
    previewProps: VerifyEmailAddress.PreviewProps,
  },
  {
    name: 'ForgotPassword',
    component: ForgotPassword,
    previewProps: ForgotPassword.PreviewProps,
  },
  {
    name: 'PasswordChanged',
    component: PasswordChanged,
    previewProps: PasswordChanged.PreviewProps,
  },
  {
    name: 'PlaidDisconnected',
    component: PlaidDisconnected,
    previewProps: PlaidDisconnected.PreviewProps,
  },
  {
    name: 'TrialAboutToExpire',
    component: TrialAboutToExpire,
    previewProps: TrialAboutToExpire.PreviewProps,
  },
];
