import { addDays, addMonths } from 'date-fns';

import { formatRelativeDate } from '@monetr/interface/util/formatDate';

import { enUS } from 'date-fns/locale';

describe('format relative date', () => {
  it('will say today and tomorrow', () => {
    expect(formatRelativeDate(new Date(), 'UTC', enUS)).toBe('today');
    expect(formatRelativeDate(addDays(new Date(), 1), 'UTC', enUS)).toBe('tomorrow');
    expect(formatRelativeDate(addDays(new Date(), -1), 'UTC', enUS)).toBe('yesterday');
  });

  it('will count days', () => {
    expect(formatRelativeDate(addDays(new Date(), 5), 'UTC', enUS)).toBe('in 5 days');
    expect(formatRelativeDate(addDays(new Date(), -5), 'UTC', enUS)).toBe('5 days ago');
  });

  it('will switch to months when its far out', () => {
    expect(formatRelativeDate(addMonths(new Date(), 4), 'UTC', enUS)).toBe('in 4 months');
  });
});
