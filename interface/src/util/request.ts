import monetrClient, { AxiosInterface } from '@monetr/interface/api/api';

export interface APIError {
  error: string;
}

export default function request(): AxiosInterface {
  return monetrClient;
}
