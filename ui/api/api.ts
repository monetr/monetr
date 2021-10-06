import axios, { AxiosInstance, AxiosRequestConfig } from "axios";

declare global {
  interface Window {
    API: AxiosInstance;
  }
}

export default axios;

export function NewClient(config: AxiosRequestConfig): AxiosInstance {
  let client = axios.create(config);

  client.interceptors.request.use(result => {
    return result;
  }, error => {
    return Promise.reject(error);
  });

  client.interceptors.response.use(result => {
    return result;
  }, error => {
    // Payment required.
    // if (error.response.status === 402) {
      // window.location.assign("/account/subscribe");
    // }

    return Promise.reject(error);
  });

  return client;
}
