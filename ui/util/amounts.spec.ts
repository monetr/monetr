import { amountToFriendly, AmountType, formatAmount, friendlyToAmount } from './amounts';

describe('amounts', () => {
  describe('amountToFriendly', () => {
    it('will convert USD cents to dollars', () => {
      const foo = amountToFriendly(1234, 'en-US', 'USD');
      expect(foo).toBe(12.34);

      const bar = amountToFriendly(1999, 'en-US', 'USD');
      expect(bar).toBe(19.99);
    });

    it('it will not clobber JPY', () => {
      const foo = amountToFriendly(1234, 'ja-JP', 'JPY');
      expect(foo).toBe(1234);

      const bar = amountToFriendly(1999, 'ja-JP', 'JPY');
      expect(bar).toBe(1999);
    });

    it('will handle more complex currencies', () => {
      const foo = amountToFriendly(1234, 'ar-BH', 'BHD');
      expect(foo).toBe(1.234);

      const bar = amountToFriendly(1999, 'ar-BH', 'BHD');
      expect(bar).toBe(1.999);
    });
  });

  describe('friendlyToAmount', () => {
    it('will convert USD cents to dollars', () => {
      const foo = friendlyToAmount(12.34, 'en-US', 'USD');
      expect(foo).toBe(1234);

      const bar = friendlyToAmount(19.99, 'en-US', 'USD');
      expect(bar).toBe(1999);
    });

    it('it will not clobber JPY', () => {
      const foo = friendlyToAmount(1234, 'ja-JP', 'JPY');
      expect(foo).toBe(1234);

      const bar = friendlyToAmount(1999, 'ja-JP', 'JPY');
      expect(bar).toBe(1999);
    });

    it('will handle more complex currencies', () => {
      const foo = friendlyToAmount(1.234, 'ar-BH', 'BHD');
      expect(foo).toBe(1234);

      const bar = friendlyToAmount(1.999, 'ar-BH', 'BHD');
      expect(bar).toBe(1999);
    });
  });

  describe('format amount', () => {
    it('will format dollar amount with defaults', () => {
      const foo = formatAmount(1234);
      expect(foo).toBe('$12.34');

      const bar = formatAmount(1001234);
      expect(bar).toBe('$10,012.34');
    });

    it('will format dollar amount with specified args', () => {
      const foo = formatAmount(1234, AmountType.Stored, 'en-US', 'USD');
      expect(foo).toBe('$12.34');
    });

    it('will format friendly amount', () => {
      const foo = formatAmount(12.34, AmountType.Friendly, 'en-US', 'USD');
      expect(foo).toBe('$12.34');
    });

    it('will format JPY properly', () => {
      const foo = formatAmount(1234, AmountType.Stored, 'ja-JP', 'JPY');
      expect(foo).toBe('￥1,234');
    });

    it('will format EUR properly', () => {
      const foo = formatAmount(1234, AmountType.Stored, 'en-DE', 'EUR');
      expect(foo).toBe('€12.34');

      const bar = formatAmount(12.34, AmountType.Friendly, 'en-DE', 'EUR');
      expect(bar).toBe('€12.34');
    });

    it('will format RUB properly', () => {
      const foo = formatAmount(1234, AmountType.Stored, 'ru-RU', 'RUB');
      expect(foo).toBe('12,34 ₽');

      const bar = formatAmount(12.34, AmountType.Friendly, 'ru-RU', 'RUB');
      expect(bar).toBe('12,34 ₽');
    });
  });
});
