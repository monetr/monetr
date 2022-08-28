import moment, { Moment } from 'moment';

import { mustParseToMoment } from 'util/parseToMoment';

describe('parse to moment', () => {
  function isMoment(input: any | Moment): input is Moment {
    return (<Moment>input).isValid();
  }

  it('will parse api date to moment', () => {
    const input = '2021-02-22T00:00:00Z';
    const result = mustParseToMoment(input);
    expect(isMoment(result)).toBeTruthy();
    expect(result.utc().format('YYYY-MM-DD')).toBe('2021-02-22');
  });

  it('will parse moment to moment', () => {
    const input = moment().startOf('day');
    const result = mustParseToMoment(input);
    expect(isMoment(result)).toBeTruthy();
    expect(result.toISOString()).toBe(input.toISOString());
  });

  it('will fail if an invalid date is provided', () => {
    const input = 'Im not a valid date.';
    expect(() => mustParseToMoment(input)).toThrow('input to mustParseToMoment was not a valid date time');
  });

  it('will fail if nothing is provided', () => {
    expect(() => mustParseToMoment(undefined)).toThrow('input to mustParseToMoment was not a valid date time');
  });
});
