import { AxiosInterface } from 'api/api';

export interface APIError {
  error: string;
}

export default function request(): AxiosInterface {
  return window.API;
}
