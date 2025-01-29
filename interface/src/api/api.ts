import * as Sentry from '@sentry/react';
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

declare global {
  interface Window {
    __MONETR__: {
      SENTRY_DSN: string | null;
    }
  }
}

export interface AxiosInterface {
  request<T = any, R = AxiosResponse<T>, D = any>(config: AxiosRequestConfig<D>): Promise<R>;
  get<T = any, R = AxiosResponse<T>, D = any>(url: string, config?: AxiosRequestConfig<D>): Promise<R>;
  delete<T = any, R = AxiosResponse<T>, D = any>(url: string, config?: AxiosRequestConfig<D>): Promise<R>;
  head<T = any, R = AxiosResponse<T>, D = any>(url: string, config?: AxiosRequestConfig<D>): Promise<R>;
  options<T = any, R = AxiosResponse<T>, D = any>(url: string, config?: AxiosRequestConfig<D>): Promise<R>;
  post<T = any, R = AxiosResponse<T>, D = any>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<R>;
  put<T = any, R = AxiosResponse<T>, D = any>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<R>;
  patch<T = any, R = AxiosResponse<T>, D = any>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<R>;
}


export function NewClient(config: AxiosRequestConfig): AxiosInstance {
  const client = axios.create(config);

  client.interceptors.request.use(result => {
    return result;
  }, error => {
    return Promise.reject(error);
  });

  client.interceptors.response.use(result => {
    return result;
  }, error => {
    // If we did get an error, and its a 500 status code then report this to sentry so we can figure out what went
    // wrong.
    if (error?.response?.status === 500) {
      Sentry.captureException(error);
    }

    // If we get an error but there is no status or response for some reason, then something else is goofy and also
    // report that to sentry so we can diagnose it.
    if (!error?.response?.status) {
      Sentry.captureException(error);
    }

    return Promise.reject(error);
  });


  return client;
}

const monetrClient = NewClient({
  baseURL: '/api',
});

export default monetrClient;
