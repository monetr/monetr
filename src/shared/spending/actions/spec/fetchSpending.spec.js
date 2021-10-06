import axios from "axios";
import Spending from "data/Spending";
import { CHANGE_BANK_ACCOUNT } from "shared/bankAccounts/actions";
import { FETCH_SPENDING_SUCCESS } from "shared/spending/actions";
import fetchSpending from "shared/spending/actions/fetchSpending";
import { mockAxiosGetOnce } from "testutils/axios";
import { createTestStore } from "testutils/store";

jest.mock('axios');

describe('fetchSpending', () => {
  it('will resolve with no bank account selected', async () => {
    const dispatch = jest.fn();
    const store = createTestStore();

    await expect(fetchSpending()(dispatch, store.getState)).resolves.toBeUndefined();

    expect(dispatch).not.toHaveBeenCalled();
  });

  it('will fetch spending successfully', async () => {
    window.API = axios.create();

    const data = {
      data: [
        new Spending({
          spendingId: 123,
          bankAccountId: 345,
          name: 'Test'
        })
      ]
    };

    const dispatch = jest.fn();
    const store = createTestStore();

    store.dispatch({
      type: CHANGE_BANK_ACCOUNT,
      bankAccountId: 345,
    });

    mockAxiosGetOnce(Promise.resolve(data));

    await expect(fetchSpending()(dispatch, store.getState)).resolves.toBeUndefined();

    expect(dispatch).toHaveBeenCalledWith({
      type: FETCH_SPENDING_SUCCESS,
      payload: expect.anything(),
    });
  });

  it('will fail to fetch spending', async () => {
    window.API = axios.create();

    const data = {
      data: {
        error: 'shits broken'
      }
    };

    const dispatch = jest.fn();
    const store = createTestStore();
    store.dispatch({
      type: CHANGE_BANK_ACCOUNT,
      bankAccountId: 345,
    });

    mockAxiosGetOnce(Promise.reject(data));

    await expect(fetchSpending()(jest.fn, store.getState)).rejects.toBe(data);

    expect(dispatch).not.toHaveBeenCalled();
  });
});

