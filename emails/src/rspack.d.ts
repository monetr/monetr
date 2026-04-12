// Ambient types for rspack/webpack require.context
interface RequireContext {
  keys(): string[];
  <T = unknown>(id: string): T;
}

// SCSS module imports
declare module '*.module.scss' {
  const classes: { readonly [key: string]: string };
  export default classes;
}
