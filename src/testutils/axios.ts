import axios from "axios";

export function mockAxiosGetOnce(result: any) {
  // @ts-ignore
  axios.get.mockImplementationOnce(() => result);
  mockAxios();
}

export function mockAxios() {
  Object.defineProperty(window, 'API', {
    value: axios
  })
}
