import React from 'react';
import { waitFor } from '@testing-library/react';

import Register from '@monetr/interface/pages/register';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('register page', () => {
  it('will render with default options', async () => {
    // server.listen();
    // server.use(
    //   http.get('/api/config', () => {
    //     return HttpResponse.json({
    //       allowSignUp: true,
    //     });
    //   }),
    // );

    const world = testRenderer(<Register />, { initialRoute: '/register' });

    await waitFor(() => expect(world.getByTestId('register-first-name')).toBeDefined());
    await waitFor(() => expect(world.getByTestId('register-last-name')).toBeDefined());
    await waitFor(() => expect(world.getByTestId('register-email')).toBeDefined());
    await waitFor(() => expect(world.getByTestId('register-password')).toBeDefined());
    await waitFor(() => expect(world.getByTestId('register-confirm-password')).toBeDefined());
    await waitFor(() => expect(world.getByTestId('register-submit')).toBeDefined());
  });
});
