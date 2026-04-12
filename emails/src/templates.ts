// PreviewProps provides sample data for the dev preview; default prop values
// contain Go template placeholders for the production build.
export type EmailTemplate = React.ComponentType<any> & {
  PreviewProps: Record<string, any>;
};

// Auto-discover all email templates via require.context at compile time.
const ctx = require.context('./emails', true, /^\.\/[^/]+\/index\.tsx$/);

export const templates: Record<string, EmailTemplate> = {};

for (const key of ctx.keys()) {
  const mod = ctx(key);
  const name = key.split('/')[1]; // "./VerifyEmailAddress/index.tsx" -> "VerifyEmailAddress"
  const Component = mod[name] || mod.default;
  if (Component && typeof Component === 'function') {
    templates[name] = Component;
  }
}
