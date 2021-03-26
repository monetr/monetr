import axios from "axios";
import Spending from "data/Spending";
import fetchSpending from "shared/spending/actions/fetchSpending";
import { mockAxios } from "testutils/axios";

jest.mock('axios');

describe('fetchSpending', () => {
  it('will fetch spending successfully', async () => {
    window.API = axios.create();

    const data = {
      data: [
        new Spending({
          spendingId: 123,
          name: 'Test'
        })
      ]
    };

    axios.get.mockImplementationOnce(() => Promise.resolve(data));
    mockAxios()

    await expect(fetchSpending()(jest.fn)).resolves.toBeUndefined();
  });

  it('will fail to fetch spending', async () => {
    window.API = axios.create();

    const data = {
      data: {
        error: 'shits broken'
      }
    };

    axios.get.mockImplementationOnce(() => Promise.reject(data));
    mockAxios()

    await expect(fetchSpending()(jest.fn)).rejects.toBe(data);
  });
});

