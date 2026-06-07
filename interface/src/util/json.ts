import type { ID, Prefixed } from '@monetr/interface/models/ID';

export type JsonEquivalent<T> = T extends Date
  ? string
  : T extends ID<Prefixed<string>>
    ? T
    : T extends Array<infer U>
      ? Array<T[number] | JsonEquivalent<U>>
      : T extends object
        ? {
            [K in keyof T as K extends symbol ? never : K]: T[K] | JsonEquivalent<T[K]>;
          }
        : T;

export type WithJsonValues<T> = {
  [K in keyof T]: T[K] | JsonEquivalent<T[K]>;
};
