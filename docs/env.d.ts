declare module '*.module.css' {
  // biome-ignore lint/suspicious/noExplicitAny: dynamic class-name keys
  const classes: any;
  export default classes;
}

declare module '*.module.scss' {
  // biome-ignore lint/suspicious/noExplicitAny: dynamic class-name keys
  const classes: any;
  export default classes;
}
