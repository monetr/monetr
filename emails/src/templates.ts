// Each template component has a PreviewProps static property with sample data
// for the dev preview, and Go template placeholders as default prop values for
// the production build.
export type EmailTemplate = React.ComponentType<any> & {
  PreviewProps: Record<string, any>;
};

// Auto-discover all email templates at compile time.
// Adding a new template: just create a new directory under src/emails/ with an
// index.tsx that exports a named component with a PreviewProps static property.
const ctx = require.context('./emails', true, /^\.\/[^/]+\/index\.tsx$/);

export const templates: Record<string, EmailTemplate> = {};

for (const key of ctx.keys()) {
  const mod = ctx(key);
  // Extract template name from path: "./VerifyEmailAddress/index.tsx" → "VerifyEmailAddress"
  const name = key.split('/')[1];
  // Use the named export matching the directory name, or fall back to default
  const Component = mod[name] || mod.default;
  if (Component && typeof Component === 'function') {
    templates[name] = Component;
  }
}
