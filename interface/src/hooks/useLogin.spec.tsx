import { act, useEffect } from 'react';

import { renderHook } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';
import NiceModal from '@ebay/nice-modal-react';
import { MemoryRouter, useLocation } from 'react-router-dom';

import monetrClient from '@monetr/interface/api/api';
import useLogin from '@monetr/interface/hooks/useLogin';
import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';

let currentLocation: ReturnType<typeof useLocation> | null = null;

function LocationSpy() {
  const location = useLocation();
  useEffect(() => {
    currentLocation = location;
  }, [location]);
  return null;
}

function wrapper({ children }: React.PropsWithChildren) {
  return (
    <MemoryRouter
      future={{ v7_startTransition: false, v7_relativeSplatPath: false }}
      initialEntries={['/login']}
    >
      <LocationSpy />
      <MQueryClient>
        <MSnackbarProvider>
          <NiceModal.Provider>{children}</NiceModal.Provider>
        </MSnackbarProvider>
      </MQueryClient>
    </MemoryRouter>
  );
}

describe('login', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
    currentLocation = null;
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will authenticate successfully', async () => {
    mockAxios.onPost('/api/authentication/login').reply(200, {
      isActive: false,
      nextUrl: '/account/subscribe',
    });

    const { result } = renderHook(useLogin, { wrapper });

    await act(() => {
      return result.current({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // Make sure we end up navigating to the url returned by the login endpoint.
    expect(currentLocation?.pathname).toBe('/account/subscribe');
  });

  it('will navigate without a next url', async () => {
    mockAxios.onPost('/api/authentication/login').reply(200, {
      isActive: true,
    });

    const { result } = renderHook(useLogin, { wrapper });

    await act(() => {
      return result.current({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When the login endpoint does not return a next url, navigate to an index route.
    expect(currentLocation?.pathname).toBe('/');
  });

  it('will require a password reset', async () => {
    mockAxios.onPost('/api/authentication/login').reply(428, {
      code: 'PASSWORD_CHANGE_REQUIRED',
      resetToken: 'abc123',
    });

    const { result } = renderHook(useLogin, { wrapper });

    await act(() => {
      return result.current({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When the login endpoint returns a password change required error; then make sure we navigate to the password
    // reset page.
    expect(currentLocation?.pathname).toBe('/password/reset');
    expect(currentLocation?.state).toEqual({
      message: 'You are required to change your password before authenticating.',
      token: 'abc123',
    });
  });

  it('email has not been verified', async () => {
    mockAxios.onPost('/api/authentication/login').reply(428, {
      code: 'EMAIL_NOT_VERIFIED',
    });

    const { result } = renderHook(useLogin, { wrapper });

    await act(() => {
      return result.current({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When our email is not verified, make sure we navigate to the resend page.
    expect(currentLocation?.pathname).toBe('/verify/email/resend');
    expect(currentLocation?.state).toEqual({
      emailAddress: 'test@test.com',
    });
  });
});
