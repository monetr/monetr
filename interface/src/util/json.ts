import type { ID, Prefixed } from '@monetr/interface/models/ID';

// JsonDataKey decides whether a given key should survive into the JSON shape of one of our models. We drop two kinds of
// keys: the symbol brand ([idPrefix]) that we use to make the ID types nominal, and any method (function valued) keys.
// Neither of those is ever part of the data the API actually sends us, so requiring them would make it impossible to
// build a model from a plain object literal or a spread of an existing instance.
type JsonDataKey<K extends PropertyKey, V> = K extends symbol
  ? never
  : V extends (...args: Array<never>) => unknown
    ? never
    : K;

export type JsonEquivalent<T> = T extends Date
  ? string
  : T extends ID<Prefixed<string>>
    ? T
    : T extends Array<infer U>
      ? Array<T[number] | JsonEquivalent<U>>
      : T extends object
        ? {
            [K in keyof T as JsonDataKey<K, T[K]>]: T[K] | JsonEquivalent<T[K]>;
          }
        : T;

// WithJsonValues takes one of our model types and produces the shape that the constructor actually receives: every
// field can be either its hydrated form or its raw JSON form (dates as strings, etc.).
export type WithJsonValues<T> = {
  [K in keyof T as JsonDataKey<K, T[K]>]: T[K] | JsonEquivalent<T[K]>;
};
