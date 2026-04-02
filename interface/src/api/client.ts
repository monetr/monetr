import { captureException } from '@sentry/react';

declare global {
  interface Window {
    __MONETR__: {
      SENTRY_DSN: string | null;
    };
  }
}

export interface RequestConfig<TRequest = unknown> {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS';
  url: string;
  data?: TRequest;
  params?: Record<string, string | number | boolean | undefined>;
  headers?: Record<string, string>;
  onUploadProgress?: (event: { loaded: number; total: number }) => void;
}

export interface ApiResponse<TResponse> {
  data: TResponse;
  status: number;
  headers: Headers;
}

export interface ApiErrorResponse<TError = unknown> {
  status: number;
  data: TError;
  headers: Headers;
}

export class ApiError<TError = unknown> extends Error {
  response: ApiErrorResponse<TError>;
  config: RequestConfig;

  constructor(message: string, config: RequestConfig, response: ApiErrorResponse<TError>) {
    super(message);
    this.name = 'ApiError';
    this.config = config;
    this.response = response;
  }
}

function buildUrl<TRequest>(config: RequestConfig<TRequest>): string {
  let url = config.url;
  if (config.params) {
    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(config.params)) {
      if (value !== undefined) {
        searchParams.append(key, String(value));
      }
    }
    const queryString = searchParams.toString();
    if (queryString) {
      url += `?${queryString}`;
    }
  }
  return url;
}

function xhrUpload<TResponse>(config: RequestConfig): Promise<ApiResponse<TResponse>> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const url = buildUrl(config);

    xhr.open(config.method, url);

    // Set any custom headers (but NOT Content-Type -- browser sets multipart boundary for FormData)
    if (config.headers) {
      for (const [key, value] of Object.entries(config.headers)) {
        xhr.setRequestHeader(key, value);
      }
    }

    // Wire up progress callback
    if (config.onUploadProgress) {
      xhr.upload.onprogress = (event: ProgressEvent) => {
        if (event.lengthComputable) {
          config.onUploadProgress({ loaded: event.loaded, total: event.total });
        }
      };
    }

    xhr.onload = () => {
      let data: TResponse;
      try {
        data = JSON.parse(xhr.responseText);
      } catch {
        data = xhr.responseText as unknown as TResponse;
      }

      const headers = new Headers();
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve({ data, status: xhr.status, headers });
      } else {
        const error = new ApiError(`Request failed with status code ${xhr.status}`, config, {
          status: xhr.status,
          data: data as unknown,
          headers,
        });
        if (xhr.status === 500) {
          captureException(error);
        }
        reject(error);
      }
    };

    xhr.onerror = () => {
      const error = new ApiError('Network Error', config, { status: 0, data: null as unknown, headers: new Headers() });
      captureException(error);
      reject(error);
    };

    xhr.send(config.data as FormData);
  });
}

export async function request<TResponse = unknown, TRequest = unknown, TError = unknown>(
  config: RequestConfig<TRequest>,
): Promise<ApiResponse<TResponse>> {
  // XHR fallback for upload progress (fetch doesn't support it)
  if (config.onUploadProgress && config.data instanceof FormData) {
    return xhrUpload<TResponse>(config);
  }

  const url = buildUrl(config);

  const init: RequestInit = {
    method: config.method,
  };

  // Set headers
  const headers: Record<string, string> = { ...config.headers };
  if (config.data !== undefined && !(config.data instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
    init.body = JSON.stringify(config.data);
  } else if (config.data instanceof FormData) {
    // Let the browser set the Content-Type with the multipart boundary
    init.body = config.data;
  }

  if (Object.keys(headers).length > 0) {
    init.headers = headers;
  }

  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (networkError) {
    const message = networkError instanceof Error ? networkError.message : 'Network Error';
    const error = new ApiError<TError>(message, config, {
      status: 0,
      data: null as unknown as TError,
      headers: new Headers(),
    });
    error.cause = networkError;
    captureException(error);
    throw error;
  }

  // Parse response body
  let data: TResponse;
  if (response.status === 204) {
    data = null as unknown as TResponse;
  } else {
    try {
      data = await response.json();
    } catch {
      data = null as unknown as TResponse;
    }
  }

  if (!response.ok) {
    const error = new ApiError<TError>(`Request failed with status code ${response.status}`, config, {
      status: response.status,
      data: data as unknown as TError,
      headers: response.headers,
    });

    if (response.status === 500) {
      captureException(error);
    }

    throw error;
  }

  return {
    data,
    status: response.status,
    headers: response.headers,
  };
}
