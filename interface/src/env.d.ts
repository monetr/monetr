declare module '*.svg' {
  const content: string;
  export default content;
}

declare module '*.module.css' {
  // biome-ignore lint/suspicious/noExplicitAny: This is just to make editors happy!
  const classes: any;
  export default classes;
}
declare module '*.module.scss' {
  // biome-ignore lint/suspicious/noExplicitAny: This is just to make editors happy!
  const classes: any;
  export default classes;
}
// declare module '*.module.sass' {
//   const classes: { readonly [key: string]: string };
//   export default classes;
// }
