import { makeFetchTransport } from '@sentry/browser';
import type { BrowserTransportOptions } from '@sentry/browser/build/npm/types/transports/types';
import type { Transport } from '@sentry/core/build/types/types-hoist/transport';

// Get the type of the second argument parameter. This way we are typesafe but we don't really need to worry about this
// parameter.
type nativeFetchType = Parameters<typeof makeFetchTransport>[1];

/**
 * @name `makeSneakyFetchTransport`
 * @description This is a wrapper around Sentry's `makeFetchTransport`. It modifies the URL and headers for the
 *              transport in order to prevent ublock origin or other browser extensions from blocking requests to sentry
 *              when a relay server is being used.
 * @returns `Transport`
 */
export function makeSneakyFetchTransport(options: BrowserTransportOptions, nativeFetch?: nativeFetchType): Transport {
  // Parse the original URL that was provided. This will have the Sentry key as a query parameter.
  const parsedUrl = new URL(options.url);
  // Take all the bits of the URL that we actually want, excluding thte query parameters.
  const newUrl = `${parsedUrl.protocol}//${parsedUrl.host}${parsedUrl.pathname}`;

  // Take the authentication out of the query params.
  const authParts: Array<string> = [];
  parsedUrl.searchParams.forEach((value, key) => authParts.push(`${key}=${value}`));

  const newOptions = {
    ...options,
    url: newUrl,
    headers: {
      ...options.headers,
      'X-Sentry-Auth': `Sentry ${authParts.join(', ')}`,
    },
  };

  return makeFetchTransport(newOptions, nativeFetch);
}
