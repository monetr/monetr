import verifyEmailAddress from 'util/verifyEmailAddress';

describe('verify email address', () => {
  it('will accept a valid email', () => {
    const address = 'test@test.com';
    expect(verifyEmailAddress(address)).toBeTruthy();
  });

  it('will accept a google alias address', () => {
    const address = 'test+alias@gmail.com';
    expect(verifyEmailAddress(address)).toBeTruthy();
  });

  it('will deny an invalid address', () => {
    const address = 'test';
    expect(verifyEmailAddress(address)).toBeFalsy();
  });
});
