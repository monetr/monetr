import { fallback } from '@monetr/interface/util/fallback';

describe('fallback', () => {
  it('will return the first value when it is defined', () => {
    expect(fallback('first', 'second')).toBe('first');
  });

  it('will skip leading undefined values', () => {
    expect(fallback(undefined, undefined, 'third', 'fourth')).toBe('third');
  });

  it('will return the final value when everything before it is undefined', () => {
    expect(fallback(undefined, undefined, 'default')).toBe('default');
  });

  it('will not treat other falsy values as undefined', () => {
    // Zero is a defined value so it should win over the trailing default.
    expect(fallback<number>(undefined, 0, 1)).toBe(0);
  });
});
