import monetrClient, { AxiosInterface } from '@monetr/interface/api/api';

export interface APIError {
  error: string;
}

/**
 * @deprecated Use axios directly instead
 */
export default function request(): AxiosInterface {
  return monetrClient;
}
