import capitalize from '@monetr/interface/util/capitalize';

describe('capitalize', () => {
  it('will capitalize the first letter', () => {
    const input = 'hello';
    const result = capitalize(input);
    expect(result).toBe('Hello');
  });

  it('will handle empty string', () => {
    const input = '';
    const result = capitalize(input);
    expect(result).toBe('');
  });

  it('will handle leading whitespace (poorly)', () => {
    const input = ' hello';
    const result = capitalize(input);
    expect(result).toBe(' hello');
  });
});
