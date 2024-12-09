import React from 'react';
import { waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import Register from '@monetr/interface/pages/register';
import testRenderer from '@monetr/interface/testutils/renderer';

import { afterAll, afterEach, beforeEach, describe, expect, it } from 'vitest';

describe('register page', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });

  afterAll(() => mockAxios.restore());

  it('will render with default options', async () => {
    mockAxios.onGet('/api/config').reply(200, {
      allowSignUp: true,
    });

    const world = testRenderer(<Register />, { initialRoute: '/register' });

    await waitFor(() => expect(world.getByTestId('register-first-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-last-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-confirm-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-submit')).toBeVisible());
  });
});
