import axios from "axios";

export function mockAxios() {
  Object.defineProperty(window, 'API', {
    value: axios
  })
}
