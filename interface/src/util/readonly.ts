export type IfEquals<X, Y, A = X, B = never> =
  (<T>() => T extends X ? 1 : 2) extends <T>() => T extends Y ? 1 : 2 ? A : B;

// WritableKeys are the keys of T that are not readonly. We also drop method (function valued) keys because they live on
// the prototype and are never something you would send back to the server, so treating them as "writable" just makes it
// impossible to build a request out of a plain object literal.
export type WritableKeys<T> = {
  [P in keyof T]-?: T[P] extends (...args: Array<never>) => unknown
    ? never
    : IfEquals<{ [Q in P]: T[P] }, { -readonly [Q in P]: T[P] }, P>;
}[keyof T];

export type Writable<T> = Pick<T, WritableKeys<T>>;
