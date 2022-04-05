import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import * as Sentry from '@sentry/react';

declare global {
  interface Window {
    API: AxiosInterface;
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
