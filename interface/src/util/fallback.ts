// fallback returns the first defined value from the provided list. The tuple type requires the final argument to always
// be defined, that way the type system knows there is always at least one value to return and we dont need a trailing
// ?? someDefault at every call site to satisfy the compiler.
export function fallback<T>(...values: [...Array<T | undefined>, T]): T {
  // The tuple type guarantees the last value is defined, so find will never actually return undefined here.
  return values.find((value): value is T => value !== undefined)!;
}
