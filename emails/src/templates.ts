/// <reference types="@rspack/core/module" />
// PreviewProps provides sample data for the dev preview; default prop values contain Go template placeholders for the
// production build.
export type EmailTemplate = React.ComponentType<unknown> & {
  PreviewProps: Record<string, string>;
};

// Auto-discover all email templates via require.context at compile time.
type TemplateModule = { [key: string]: EmailTemplate | undefined; default?: EmailTemplate };
const ctx = require.context('./emails', false, /^\.\/[^/]+\.tsx$/) as {
  keys(): string[];
  (key: string): TemplateModule;
};

export const templates: Record<string, EmailTemplate> = {};

for (const key of ctx.keys()) {
  const mod = ctx(key);
  const name = key.slice(2).replace(/\.tsx$/, ''); // "./VerifyEmailAddress.tsx" -> "VerifyEmailAddress"
  const Component = mod[name] || mod.default;
  if (Component && typeof Component === 'function') {
    templates[name] = Component;
  }
}
