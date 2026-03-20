import { useCallback, useEffect, useRef, useState } from 'react';

const DEFAULT_DELAY_IN_MS = 100;

interface UseInfiniteScrollOptions {
  loading: boolean;
  hasNextPage: boolean;
  onLoadMore: () => void;
  disabled?: boolean;
  rootMargin?: string;
  delayInMs?: number;
}

export function useInfiniteScroll({
  loading,
  hasNextPage,
  onLoadMore,
  disabled = false,
  rootMargin = '0px',
  delayInMs = DEFAULT_DELAY_IN_MS,
}: UseInfiniteScrollOptions): [(node: Element | null) => void] {
  const [isVisible, setIsVisible] = useState(false);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const rootMarginRef = useRef(rootMargin);
  rootMarginRef.current = rootMargin;

  const refCallback = useCallback((node: Element | null) => {
    if (observerRef.current) {
      observerRef.current.disconnect();
      observerRef.current = null;
    }

    if (!node) {
      setIsVisible(false);
      return;
    }

    const observer = new IntersectionObserver(
      entries => {
        const entry = entries[0];
        if (entry) {
          setIsVisible(entry.isIntersecting);
        }
      },
      { rootMargin: rootMarginRef.current },
    );

    observer.observe(node);
    observerRef.current = observer;
  }, []);

  useEffect(() => {
    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, []);

  const shouldLoadMore = !disabled && !loading && isVisible && hasNextPage;

  useEffect(() => {
    if (!shouldLoadMore) {
      return undefined;
    }

    const timer = setTimeout(() => {
      onLoadMore();
    }, delayInMs);

    return () => clearTimeout(timer);
  }, [shouldLoadMore, onLoadMore, delayInMs]);

  return [refCallback];
}
