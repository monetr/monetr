export declare const idPrefix: unique symbol;

export interface Prefixed<P extends string> {
  readonly [idPrefix]: P;
}

export type PrefixOf<T> = T extends Prefixed<infer P> ? P : never;

export type ID<T extends Prefixed<string>> = `${PrefixOf<T>}_${string}` & { readonly [idPrefix]: T };

export const ID = {
  isZero<T extends Prefixed<string>>(id: ID<T> | null | undefined): boolean {
    if (!id) {
      return true;
    }
    const i = id.indexOf('_');
    return i === -1 ? id.length === 0 : i === id.length - 1;
  },

  prefix<T extends Prefixed<string>>(id: ID<T>): PrefixOf<T> {
    return id.slice(0, id.indexOf('_')) as PrefixOf<T>;
  },

  withoutPrefix<T extends Prefixed<string>>(id: ID<T>): string {
    return id.slice(id.indexOf('_') + 1);
  },

  /**
   * `from` takes a string input and converts it into an ID, but if the string is free form input in code, it will
   * enforce that the string properly contains the expected prefix.
   *
   * If you receive a `never` type error then your string input is not valid for the ID you are trying to use.
   */
  from<T extends Prefixed<string>, S extends string = `${PrefixOf<T>}_${string}`>(value: S): ID<T> {
    // Evil casting :(
    return value as unknown as ID<T>;
  },

  parse<T extends Prefixed<string>>(value: string, expectedPrefix: PrefixOf<T>): ID<T> {
    const p = `${expectedPrefix}_`;
    if (!value.startsWith(p) || value.length === p.length) {
      throw new Error(`invalid ID, expected prefix "${p}", got "${value}"`);
    }
    return value as unknown as ID<T>;
  },
};
