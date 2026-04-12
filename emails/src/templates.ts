/// <reference types="@rspack/core/module" />
// PreviewProps provides sample data for the dev preview; default prop values contain Go template placeholders for the
// production build.
export type EmailTemplate = React.ComponentType<unknown> & {
  PreviewProps: Record<string, string>;
};

// Auto-discover all email templates via require.context at compile time.
const ctx = require.context('./emails', false, /^\.\/[^/]+\.tsx$/);
export const templates: Record<string, EmailTemplate> = {};

for (const key of ctx.keys()) {
  const module = ctx(key);
  const name = key.slice(2).replace(/\.tsx$/, ''); // "./VerifyEmailAddress.tsx" -> "VerifyEmailAddress"
  const Component = module[name];
  if (Component && typeof Component === 'function') {
    templates[name] = Component;
  }
}
