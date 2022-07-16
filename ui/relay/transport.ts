import { BaseTransport } from '@sentry/browser/dist/transports';
import { APIDetails, eventToSentryRequest, sessionToSentryRequest } from '@sentry/core';
import {
  Event,
  Response as SentryResponse,
  SentryRequest,
  SentryRequestType,
  Session,
  SessionAggregates,
  TransportOptions,
} from '@sentry/types';
import {
  eventStatusFromHttpCode,
  isRateLimited,
  logger,
  SentryError,
  SyncPromise,
  updateRateLimits } from '@sentry/utils';
import axios, { AxiosResponse } from 'axios';

interface SentryRequestExtended extends SentryRequest {
  headers: { [header: string]: string };
}

export default class RelayTransport extends BaseTransport {

  public constructor(public options: TransportOptions) {
    super(options);
  }

  sendEvent(event: Event): PromiseLike<SentryResponse> {
    return this._sendRequest(this._eventToSentryRequest(event, this._api), event);
  }

  sendSession(session: Session | SessionAggregates): PromiseLike<SentryResponse> {
    return this._sendRequest(this._sessionToSentryRequest(session, this._api), session);
  }

  private _sessionToSentryRequest(session: Session | SessionAggregates, api: APIDetails): SentryRequestExtended {
    return this._extendSentryRequest(sessionToSentryRequest(session, api));
  }

  private _eventToSentryRequest(event: Event, api: APIDetails): SentryRequestExtended {
    return this._extendSentryRequest(eventToSentryRequest(event, api));
  }

  private _extendSentryRequest(request: SentryRequest): SentryRequestExtended {
    const parsedUrl = new URL(request.url);
    const newUrl = `${ parsedUrl.protocol }//${ parsedUrl.host }${ parsedUrl.pathname }`;

    const authParts: string[] = [];
    parsedUrl.searchParams.forEach((value, key) => authParts.push(`${ key }=${ value }`));

    return {
      body: request.body,
      headers: {
        'X-Sentry-Auth': `Sentry ${ authParts.join(', ') }`,
      },
      type: request.type,
      url: newUrl,
    };
  }

  _sendRequest(sentryRequest: SentryRequestExtended, originalPayload: Event | Session | SessionAggregates): PromiseLike<SentryResponse> {
    // Check if the current request type is being rate limited.
    if (isRateLimited(this._rateLimits, sentryRequest.type)) {
      // If it is then we want to record a lost event with the rate limit backoff outcome for this request type.
      this.recordLostEvent('ratelimit_backoff', sentryRequest.type);

      // Then return a rejected promise to the caller indicating that this request will not be sent.
      return Promise.reject({
        event: originalPayload,
        type: sentryRequest.type,
        reason: `Transport for ${ sentryRequest.type } requests locked till ${ this._disabledUntil(
          sentryRequest.type,
        ) } due to too many requests.`,
        status: 429,
      });
    }

    return this._buffer.add(
      () => new SyncPromise<SentryResponse>((resolve, reject) => {
        axios.post<SentryResponse>(sentryRequest.url, sentryRequest.body, {
          headers: sentryRequest.headers,
        })
          .then(response => {
            const headers = {
              'x-sentry-rate-limits': response.headers['X-Sentry-Rate-Limits'],
              'retry-after': response.headers['Retry-After'],
            };
            this._handleAxiosResponse(
              sentryRequest.type,
              response,
              headers,
              resolve,
              reject,
            );
          }).catch(reject);
      }),
    )
      .then(undefined, reason => {
        // It's either buffer rejection or any other xhr/fetch error, which are treated as NetworkError.
        if (reason instanceof SentryError) {
          this.recordLostEvent('queue_overflow', sentryRequest.type);
        } else {
          this.recordLostEvent('network_error', sentryRequest.type);
        }
        throw reason;
      });
  }

  protected _handleAxiosResponse(
    requestType: SentryRequestType,
    response: AxiosResponse,
    headers: Record<string, string | null>,
    resolve: (value?: SentryResponse | PromiseLike<SentryResponse> | null | undefined) => void,
    reject: (reason?: unknown) => void,
  ): void {
    const status = eventStatusFromHttpCode(response.status);

    /**
     * "The name is case-insensitive."
     * https://developer.mozilla.org/en-US/docs/Web/API/Headers/get
     */
    this._rateLimits = updateRateLimits(this._rateLimits, headers);
    if (isRateLimited(this._rateLimits, requestType))
      logger.warn(`Too many ${ requestType } requests, backing off until: ${ this._disabledUntil(requestType) }`);

    if (status === 'success') {
      resolve({ status });
      return;
    }

    reject(response);
  }
}
