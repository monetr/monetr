import formatAmount from 'util/formatAmount';

describe('formatAmount', () => {
  it('will format dollar amount', () => {
    const result = formatAmount(1234);
    expect(result).toBe('$12.34');
  });
});
