export {
  ApiError,
  type ApiResponse,
  type RequestConfig,
  request as default,
} from '@monetr/interface/api/client';

export interface APIError {
  error: string;
  problems?: { [key: string]: string };
}
