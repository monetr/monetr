import { act } from 'react';

import { waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import SettingsAPIKeys from '@monetr/interface/pages/settings/api';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('api keys settings page', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    // Every state of the page needs the config (the create modal reads proof of work), the authenticated user (locale
    // and timezone for the key items) and the user lookup for the Created By line.
    mockFetch.onGet('/api/config').reply(200, {
      allowSignUp: true,
      billingEnabled: false,
      proofOfWorkEnabled: false,
    });
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: null,
      hasSubscription: false,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      mfaPending: false,
      user: {
        userId: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
        loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
        login: {
          loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2025-01-07T18:39:50.227236Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
        account: {
          accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
          timezone: 'America/Chicago',
          locale: 'en_US',
          createdAt: '2025-01-07T18:39:46.40702Z',
        },
        role: 'owner',
      },
    });
    mockFetch.onGet('/api/users/user_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      userId: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
      loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
      login: {
        loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
        email: 'example@example.com',
        firstName: 'Elliot',
        lastName: 'Courant',
        passwordResetAt: null,
        isEmailVerified: true,
        emailVerifiedAt: '2025-01-07T18:39:50.227236Z',
        totpEnabledAt: null,
      },
      accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
      account: {
        accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
        timezone: 'America/Chicago',
        locale: 'en_US',
        createdAt: '2025-01-07T18:39:46.40702Z',
      },
    });
  });
  afterEach(() => {
    mockFetch.reset();
  });

  afterAll(() => {
    mockFetch.restore();
  });

  it('will show the empty state when there are no keys', async () => {
    mockFetch.onGet('/api/keys').reply(200, []);

    const world = testRenderer(<SettingsAPIKeys />, { initialRoute: '/settings/api' });

    await waitFor(() => expect(world.getByTestId('api-keys-empty')).toBeVisible());
    expect(world.getByText('No API Keys yet')).toBeVisible();
  });

  it('will list the api keys for the account', async () => {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc',
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6xyz',
        name: 'CI Deploys',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderer(<SettingsAPIKeys />, { initialRoute: '/settings/api' });

    // Each key card carries the key's id as its DOM id, making individual keys directly addressable.
    await waitFor(() => expect(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')).toBeVisible());
    await waitFor(() => expect(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6xyz')).toBeVisible());
    expect(world.getByText('Personal Automation')).toBeVisible();
    expect(world.getByText('CI Deploys')).toBeVisible();
    // The Created By line resolves the creator via /api/users/{userId}.
    await waitFor(() => expect(world.getAllByText('Elliot Courant')).toHaveLength(2));
  });

  it('will create an api key and show the secret exactly once', async () => {
    mockFetch.onGet('/api/keys').reply(200, []);
    mockFetch.onPost('/api/keys').reply(200, {
      apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc',
      name: 'CI Robot',
      createdAt: '2023-07-02T04:22:52.48118Z',
      createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
      updatedAt: '2023-07-02T04:22:52.48118Z',
      deletedAt: null,
      secret: 'monetr_secret_aebagbafaydqqcikbmga2dqpcaireeyuculbogazdinryhi6d4qa',
    });

    const world = testRenderer(<SettingsAPIKeys />, { initialRoute: '/settings/api' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('api-keys-empty-create')).toBeVisible());
    await user.click(world.getByTestId('api-keys-empty-create'));

    // The name step should be showing, fill it out and submit it.
    await waitFor(() => expect(world.getByTestId('create-api-key-modal')).toBeVisible());
    await user.type(world.getByTestId('create-api-key-name'), 'CI Robot');
    await user.click(world.getByTestId('create-api-key-submit'));

    // Once the key is created the modal should move to the secret step, showing the key id and the secret.
    await waitFor(() => expect(world.getByTestId('create-api-key-secret')).toBeVisible());
    expect(world.getByText('key_01hy4rfqk8z4xv1c2v44cf6abc')).toBeVisible();
    expect(world.getByText('monetr_secret_aebagbafaydqqcikbmga2dqpcaireeyuculbogazdinryhi6d4qa')).toBeVisible();

    // Proof of work is disabled in the config above, so the only thing that should go over the wire is the name.
    const postHistory = mockFetch.history.post;
    expect(postHistory).toHaveLength(1);
    expect(postHistory?.[0]?.data).toEqual({ name: 'CI Robot' });

    // Done closes the modal for good, the secret is never shown again.
    await user.click(world.getByTestId('close-create-api-key-secret'));
    await waitFor(() => expect(world.queryByTestId('create-api-key-secret')).not.toBeInTheDocument());
  });

  it('will revoke an api key', async () => {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc',
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);
    mockFetch.onDelete('/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc').reply(200);

    const world = testRenderer(<SettingsAPIKeys />, { initialRoute: '/settings/api' });
    const user = userEvent.setup();

    await waitFor(() => expect(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')).toBeVisible());
    const item = within(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')!);
    await user.click(item.getByRole('button', { name: 'Revoke' }));

    // The confirmation modal shows the key being revoked (without its own revoke button) before anything is sent.
    await waitFor(() => expect(world.getByTestId('revoke-api-key-modal')).toBeVisible());
    expect(mockFetch.history.delete).toHaveLength(0);

    await act(() => user.click(world.getByTestId('revoke-api-key-confirm')));

    // The modal goes away, the key was deleted server side and pruned from the cached list, leaving the empty state.
    await waitFor(() => expect(world.queryByTestId('revoke-api-key-modal')).not.toBeInTheDocument());
    const deleteHistory = mockFetch.history.delete;
    expect(deleteHistory).toHaveLength(1);
    expect(deleteHistory?.[0]).toMatchObject({ url: '/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc' });
    await waitFor(() => expect(world.getByTestId('api-keys-empty')).toBeVisible());
    expect(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')).toBeNull();
  });

  it('will keep the key when the revoke is cancelled', async () => {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc',
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk',
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderer(<SettingsAPIKeys />, { initialRoute: '/settings/api' });
    const user = userEvent.setup();

    await waitFor(() => expect(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')).toBeVisible());
    const item = within(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')!);
    await user.click(item.getByRole('button', { name: 'Revoke' }));

    await waitFor(() => expect(world.getByTestId('revoke-api-key-modal')).toBeVisible());
    await user.click(world.getByTestId('close-revoke-api-key-modal'));

    // Nothing was sent and the key is still listed.
    await waitFor(() => expect(world.queryByTestId('revoke-api-key-modal')).not.toBeInTheDocument());
    expect(mockFetch.history.delete).toHaveLength(0);
    expect(document.getElementById('key_01hy4rfqk8z4xv1c2v44cf6abc')).toBeVisible();
  });
});
