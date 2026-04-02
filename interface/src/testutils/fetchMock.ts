export interface HistoryEntry {
  url: string;
  data?: unknown;
}

interface Handler {
  method: string;
  url: string;
  status: number;
  body: unknown;
}

interface ReplyChain {
  reply(status: number, body?: unknown): FetchMock;
}

export default class FetchMock {
  private handlers: Array<Handler> = [];
  private _history: Record<string, HistoryEntry[]> = {
    get: [],
    post: [],
    put: [],
    patch: [],
    delete: [],
  };
  private originalFetch: typeof globalThis.fetch;

  constructor() {
    this.originalFetch = globalThis.fetch;
    globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
      const rawUrl = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      const method = (init?.method || 'GET').toUpperCase();

      // Record in history (strip origin to keep relative path for test assertions)
      const relativeUrl = rawUrl.startsWith('http') ? new URL(rawUrl).pathname : rawUrl;
      let data: unknown;
      if (init?.body) {
        try {
          data = JSON.parse(init.body as string);
        } catch {
          data = init.body;
        }
      }
      const methodKey = method.toLowerCase();
      if (this._history[methodKey]) {
        this._history[methodKey].push({ url: relativeUrl, data });
      }

      // Find matching handler
      const handlerIndex = this.handlers.findIndex(h => h.method === method && h.url === rawUrl);
      if (handlerIndex === -1) {
        return Promise.reject(new Error(`No mock handler for ${method} ${rawUrl}`));
      }

      const handler = this.handlers[handlerIndex];

      // Return a real Response object
      return new Response(
        handler.body !== undefined ? JSON.stringify(handler.body) : null,
        {
          status: handler.status,
          headers: { 'Content-Type': 'application/json' },
        },
      );
    };
  }

  get history() {
    return this._history;
  }

  onGet(url: string): ReplyChain {
    return this._on('GET', url);
  }

  onPost(url: string): ReplyChain {
    return this._on('POST', url);
  }

  onPut(url: string): ReplyChain {
    return this._on('PUT', url);
  }

  onPatch(url: string): ReplyChain {
    return this._on('PATCH', url);
  }

  onDelete(url: string): ReplyChain {
    return this._on('DELETE', url);
  }

  private _on(method: string, url: string): ReplyChain {
    return {
      reply: (status: number, body?: unknown): FetchMock => {
        this.handlers.push({ method, url, status, body });
        return this;
      },
    };
  }

  reset(): void {
    this.handlers = [];
    for (const key of Object.keys(this._history)) {
      this._history[key] = [];
    }
  }

  restore(): void {
    globalThis.fetch = this.originalFetch;
  }
}
