import avatarLetter from '@monetr/interface/util/avatarLetter';

describe('avatar letter', () => {
  it('will use the first character', () => {
    expect(avatarLetter('amazon')).toBe('A');
  });

  it('will skip leading symbols', () => {
    expect(avatarLetter('- Transfer')).toBe('T');
    expect(avatarLetter('#1 Pizza')).toBe('1');
  });

  it('will handle other scripts', () => {
    expect(avatarLetter('「セブン」')).toBe('セ');
    expect(avatarLetter('-éclair')).toBe('É');
  });

  it('will fall back', () => {
    expect(avatarLetter('---')).toBe('?');
    expect(avatarLetter('')).toBe('?');
    expect(avatarLetter(undefined)).toBe('?');
  });
});
