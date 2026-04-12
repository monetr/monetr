// Ambient types for rspack/webpack require.context
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
