import React from 'react';
import { waitFor } from '@testing-library/react';

import { rest } from 'msw';
import Register from 'pages/register';
import testRenderer from 'testutils/renderer';
import { server } from 'testutils/server';

describe('register page', () => {
  it('will render with default options', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowSignUp: true,
        }));
      }),
    );

    const world = testRenderer(<Register />, { initialRoute: '/register' });

    await waitFor(() => expect(world.getByTestId('register-first-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-last-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-confirm-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-submit')).toBeVisible());
  });
});
