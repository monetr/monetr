// setupTests mocks matchMedia to always return false, which means useIsMobile is always false too. This swaps in one
// that can be changed during a test. Reduced motion defaults to true so animations snap and tests don't need to wait
// on them.
export interface MatchMediaOptions {
  mobile: boolean;
  reducedMotion?: boolean;
}

export interface MatchMediaControl {
  setMobile(next: boolean): void;
  restore(): void;
}

export function installMatchMedia(options: MatchMediaOptions): MatchMediaControl {
  const original = window.matchMedia;
  const listeners = new Map<string, Set<(event: MediaQueryListEvent) => void>>();
  const state = {
    mobile: options.mobile,
    reducedMotion: options.reducedMotion ?? true,
  };

  const matchesQuery = (query: string): boolean => {
    if (query.includes('prefers-reduced-motion')) {
      return state.reducedMotion;
    }
    if (query.includes('max-width')) {
      return state.mobile;
    }
    return false;
  };

  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: (query: string) => ({
      media: query,
      get matches() {
        return matchesQuery(query);
      },
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: (_: string, listener: (event: MediaQueryListEvent) => void) => {
        const existing = listeners.get(query) ?? new Set();
        existing.add(listener);
        listeners.set(query, existing);
      },
      removeEventListener: (_: string, listener: (event: MediaQueryListEvent) => void) => {
        listeners.get(query)?.delete(listener);
      },
      dispatchEvent: () => true,
    }),
  });

  return {
    setMobile(next: boolean) {
      state.mobile = next;
      for (const [query, set] of listeners) {
        if (!query.includes('max-width')) {
          continue;
        }
        for (const listener of set) {
          listener({
            matches: next,
            media: query,
          } as MediaQueryListEvent);
        }
      }
    },
    restore() {
      Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: original,
      });
    },
  };
}
