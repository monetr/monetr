import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import * as Sentry from '@sentry/react';

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
    if (error.response.status === 500) {
      Sentry.captureException(error);
    }

    return Promise.reject(error);
  });

  return client;
}
