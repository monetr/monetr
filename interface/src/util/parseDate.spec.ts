import parseDate from '@monetr/interface/util/parseDate';

describe('parse dates', () => {
  it('will parse a date object', () => {
    const input = new Date();
    const result = parseDate(input);
    expect(result).toEqual(input);
    expect(result).toBeInstanceOf(Date);
  });

  it('will parse a date string', () => {
    const input = '2022-09-25T00:40:13.621942Z';
    const result = parseDate(input);
    expect(result).toBeInstanceOf(Date);
  });

  it('will handle a null value', () => {
    const input = null;
    const result = parseDate(input);
    expect(result).toBeNull();
  });

  it('will handle an undefined value', () => {
    const input = undefined;
    const result = parseDate(input);
    expect(result).toBeNull();
  });
});
