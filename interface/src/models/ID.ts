declare const idBrand: unique symbol;

export type ID<T> = string & { readonly [idBrand]: T };

export const ID = {
  isZero<T>(id: ID<T> | null | undefined): boolean {
    if (!id) {
      return true;
    }
    const i = id.indexOf('_');
    return i === -1 ? id.length === 0 : i === id.length - 1;
  },

  prefix<T>(id: ID<T>): string {
    const i = id.indexOf('_');
    return i === -1 ? '' : id.slice(0, i);
  },

  withoutPrefix<T>(id: ID<T>): string {
    const i = id.indexOf('_');
    return i === -1 ? id : id.slice(i + 1);
  },

  // Unchecked construction — for JSON deserialization at trust boundaries.
  from<T>(value: string): ID<T> {
    return value as ID<T>;
  },

  // Checked construction — verifies the prefix matches.
  parse<T>(value: string, expectedPrefix: string): ID<T> {
    const p = `${expectedPrefix}_`;
    if (!value.startsWith(p) || value.length === p.length) {
      throw new Error(`invalid ID, expected prefix "${p}", got "${value}"`);
    }
    return value as ID<T>;
  },
};
