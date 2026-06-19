import { act } from 'react';

import { type PatchTransactionResponse, usePatchTransaction } from '@monetr/interface/hooks/usePatchTransaction';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Transaction from '@monetr/interface/models/Transaction';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('patch transaction', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
  });
  afterEach(() => {
    mockFetch.reset();
  });

  afterAll(() => {
    mockFetch.restore();
  });

  // A reasonably complete transaction body that the server would send back. The hook does not hydrate this into a model
  // so the fields come straight back as JSON, but onSuccess does read transactionId and bankAccountId off of it.
  const transactionResponse = {
    transactionId: 'txn_01hy4rh9m2n3p4q5r6s7t8u9vw',
    bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
    amount: 1000,
    spendingId: null,
    spendingAmount: null,
    categories: ['Other'],
    date: '2023-07-31T05:00:00Z',
    authorizedDate: null,
    name: 'Updated Name',
    originalName: 'Some Store',
    merchantName: null,
    originalMerchantName: null,
    isPending: false,
    createdAt: '2023-07-02T04:22:52.48118Z',
  };

  const balanceResponse = {
    bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
    available: 48635,
    current: 48635,
    free: 40000,
    expenses: 8635,
    goals: 0,
  };

  it('will patch a transaction', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions/txn_01hy4rh9m2n3p4q5r6s7t8u9vw')
      .reply(200, {
        transaction: transactionResponse,
        spending: [],
        balance: balanceResponse,
      });

    const world = testRenderHook(usePatchTransaction, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
    });
    let result!: PatchTransactionResponse;
    await act(async () => {
      result = await world.result.current({
        transactionId: ID.from<Transaction>('txn_01hy4rh9m2n3p4q5r6s7t8u9vw'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Updated Name',
        spendingId: null,
      });
    });

    expect(result).toBeDefined();
    expect(result.transaction.transactionId).toBe('txn_01hy4rh9m2n3p4q5r6s7t8u9vw');
    expect(result.transaction.name).toBe('Updated Name');
    expect(result.balance.available).toBe(48635);
    expect(result.spending.length).toBe(0);
  });

  it('it will fail to patch a transaction', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions/txn_01hy4rh9m2n3p4q5r6s7t8u9vw')
      .reply(400, {
        error: 'Invalid request',
      });

    const world = testRenderHook(usePatchTransaction, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
    });
    await act(async () => {
      await expect(
        world.result.current({
          transactionId: ID.from<Transaction>('txn_01hy4rh9m2n3p4q5r6s7t8u9vw'),
          bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
          name: 'Updated Name',
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
        response: {
          data: {
            error: 'Invalid request',
          },
        },
      });
    });
  });

  it('will drop undefined fields from the request body but keep nulls', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions/txn_01hy4rh9m2n3p4q5r6s7t8u9vw')
      .reply(200, {
        transaction: transactionResponse,
        spending: [],
        balance: balanceResponse,
      });

    const world = testRenderHook(usePatchTransaction, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
    });
    await act(async () => {
      await world.result.current({
        transactionId: ID.from<Transaction>('txn_01hy4rh9m2n3p4q5r6s7t8u9vw'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        // name is left undefined on purpose, we did not touch it so it should not be sent at all.
        name: undefined,
        // spendingId is explicitly null, the user is clearing the expense so it MUST survive into the request body. The
        // server only clears the spending if the key is actually present as null.
        spendingId: null,
      });
    });

    // Grab the body that actually went over the wire so we can prove the undefined fields never get sent. The history
    // is keyed by method so typescript thinks the patch bucket might be undefined, pull it into a local and assert its
    // there before we poke at the first entry.
    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    const body = patchHistory?.[0]?.data as Record<string, unknown>;
    expect('name' in body).toBe(false);
    expect('spendingId' in body).toBe(true);
    expect(body.spendingId).toBeNull();
    // transactionId and bankAccountId are path params, the hook destructures them out so they should never end up in
    // the patch body.
    expect('transactionId' in body).toBe(false);
    expect('bankAccountId' in body).toBe(false);
  });

  it('will send the manual transaction fields and serialize the date', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions/txn_01hy4rh9m2n3p4q5r6s7t8u9vw')
      .reply(200, {
        transaction: transactionResponse,
        spending: [],
        balance: balanceResponse,
      });

    const world = testRenderHook(usePatchTransaction, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
    });
    // On a manual link the details page also sends the amount, date and pending state. The date goes up as a real Date
    // object so we want to make sure it gets serialized into an ISO string the server can actually parse.
    const date = new Date('2023-07-31T05:00:00.000Z');
    await act(async () => {
      await world.result.current({
        transactionId: ID.from<Transaction>('txn_01hy4rh9m2n3p4q5r6s7t8u9vw'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Updated Name',
        amount: 1000,
        date: date,
        isPending: true,
      });
    });

    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    const body = patchHistory?.[0]?.data as Record<string, unknown>;
    expect(body.amount).toBe(1000);
    expect(body.isPending).toBe(true);
    expect(body.date).toBe('2023-07-31T05:00:00.000Z');
  });
});
