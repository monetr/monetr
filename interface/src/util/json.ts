import type { ID } from '@monetr/interface/models/ID';

export type JsonEquivalent<T> = T extends Date | ID<unknown>
  ? string
  : T extends Array<infer U>
    ? Array<T[number] | JsonEquivalent<U>>
    : T extends object
      ? { [K in keyof T]: T[K] | JsonEquivalent<T[K]> }
      : T;

export type WithJsonValues<T> = {
  [K in keyof T]: T[K] | JsonEquivalent<T[K]>;
};
