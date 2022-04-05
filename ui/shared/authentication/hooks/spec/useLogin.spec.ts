import { createMemoryHistory, MemoryHistory } from 'history';
import mockAxios from 'jest-mock-axios';
import useLogin from 'shared/authentication/hooks/useLogin';
import testRenderHook from 'testutils/hooks';

describe('useLogin', () => {
  let history: MemoryHistory;
  beforeEach(() => {
    history = createMemoryHistory();
    history.push = jest.fn();
  });

  it('will handle an API failure', async () => {
    const { result: { current: login } } = testRenderHook(() => useLogin(), {
      history,
    });
    const result = login({
      email: 'email@test.com',
      password: 'iAmAPassword',
    });
    expect(mockAxios.post).toHaveBeenCalledWith('/authentication/login', {
      email: 'email@test.com',
      password: 'iAmAPassword',
    });
    mockAxios.mockError({
      response: {
        status: 500,
        data: {
          error: 'Something has gone the not right way...',
        }
      }
    });
    await expect(result).rejects.toStrictEqual({
      isAxiosError: true,
      response: { data: { error: 'Something has gone the not right way...' }, status: 500 }
    });
    expect(history.push).not.toHaveBeenCalled();
  });

  it('will handle invalid credentials', async () => {
    const { result: { current: login } } = testRenderHook(() => useLogin(), {
      history,
    });
    const result = login({
      email: 'email@test.com',
      password: 'iAmAPassword',
    });
    expect(mockAxios.post).toHaveBeenCalledWith('/authentication/login', {
      email: 'email@test.com',
      password: 'iAmAPassword',
    });
    mockAxios.mockError({
      response: {
        status: 403,
        data: {
          error: 'Invalid email or password.',
        }
      }
    });
    await expect(result).rejects.toStrictEqual({
      isAxiosError: true,
      response: { data: { error: 'Invalid email or password.' }, status: 403 }
    });
    expect(history.push).not.toHaveBeenCalled();
  });

  it('will handle email not verified', async () => {
    const { result: { current: login } } = testRenderHook(() => useLogin(), {
      history,
    });
    const result = login({
      email: 'email@test.com',
      password: 'iAmAPassword',
    });
    expect(mockAxios.post).toHaveBeenCalledWith('/authentication/login', {
      email: 'email@test.com',
      password: 'iAmAPassword',
    });
    mockAxios.mockError({
      response: {
        status: 428,
        data: {
          code: 'EMAIL_NOT_VERIFIED',
          error: 'Email address is not verified.',
        }
      }
    });
    await expect(result).resolves.toBe(undefined);
    expect(history.push).toHaveBeenCalledWith({
      hash: '',
      pathname: '/verify/email/resend',
      search: ''
    }, { emailAddress: 'email@test.com' });
  });
});
