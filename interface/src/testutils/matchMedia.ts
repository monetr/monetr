// setupTests mocks matchMedia to `matches: false` for every query, which leaves useIsMobile permanently false. This
// swaps in a query aware version for the length of a test.
//
// Reduced motion defaults to true so animated components snap and their state, focus and ARIA assertions are all
// synchronous. Tests that need the animation should pass `reducedMotion: false`.
export interface MatchMediaControl {
  setMobile(next: boolean): void;
  restore(): void;
}

export function installMatchMedia(options: { mobile: boolean; reducedMotion?: boolean }): MatchMediaControl {
  const original = window.matchMedia;
  const listeners = new Map<string, Set<(event: MediaQueryListEvent) => void>>();
  const state = { mobile: options.mobile, reducedMotion: options.reducedMotion ?? true };

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
          listener({ matches: next, media: query } as MediaQueryListEvent);
        }
      }
    },
    restore() {
      Object.defineProperty(window, 'matchMedia', { configurable: true, writable: true, value: original });
    },
  };
}
