import axios from "axios";
import Spending from "data/Spending";
import { FETCH_SPENDING_SUCCESS } from "shared/spending/actions";
import fetchSpending from "shared/spending/actions/fetchSpending";
import { mockAxiosGetOnce } from "testutils/axios";

jest.mock('axios');

describe('fetchSpending', () => {
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

    mockAxiosGetOnce(Promise.resolve(data));

    await expect(fetchSpending()(dispatch)).resolves.toBeUndefined();

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

    mockAxiosGetOnce(Promise.reject(data));

    await expect(fetchSpending()(jest.fn)).rejects.toBe(data);
  });
});

