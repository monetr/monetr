import { act } from '@testing-library/react-hooks';
import { rest } from 'msw';

import useLogout from '@monetr/interface/hooks/useLogout';
import testRenderHook from '@monetr/interface/testutils/hooks';
import { server } from '@monetr/interface/testutils/server';

describe('logout', () => {
  it('will logout successfully', async () => {
    server.use(
      rest.get('/api/authentication/logout', (_req, res, ctx) => {
        expect(_req).toBeDefined();
        return res(ctx.status(200));
      }),
    );

    const { result: { current: logout } } = testRenderHook(useLogout, { initialRoute: '/' });

    await act(() => {
      return logout();
    });
    // This test is really dumb? It basically just adds code coverage lol.
    // The real logout endpoint just removes a cookie, the redirect from logging out happens separately from this hook.
    // Just make sure that we did actually call the endpoint.
    expect.assertions(1);
  });
});
