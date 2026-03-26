// Type declarations for rspack/webpack's require.context, used to auto-discover
// email templates at compile time.
interface RequireContext {
  keys(): string[];
  <T = any>(id: string): T;
}

interface NodeRequire {
  context(directory: string, useSubdirectories: boolean, regExp: RegExp): RequireContext;
}

// SCSS module imports
declare module '*.module.scss' {
  const classes: { readonly [key: string]: string };
  export default classes;
}
