import React from 'react';
import { useLocation } from 'react-router-dom';

interface ScrollPositionProviderProps {
  children: React.ReactNode;
}

interface IScrollPositionContext {
  saveScrollPosition: (key: string, position: number) => void;
  getScrollPosition: (key: string) => number;
}

const ScrollPositionContext = React.createContext<
  IScrollPositionContext | undefined
>(undefined);

function useScrollPosition(): IScrollPositionContext {
  const context = React.useContext(ScrollPositionContext);
  if (!context) {
    throw new Error(
      'useScrollPosition must be used within a ScrollPositionProvider',
    );
  }
  return context;
}

export function ScrollPositionProvider({
  children,
}: ScrollPositionProviderProps): JSX.Element {
  const scrollPositionsRef = React.useRef<{ [key: string]: number }>({});

  const saveScrollPosition = React.useCallback(
    (key: string, position: number) => {
      scrollPositionsRef.current = {
        ...scrollPositionsRef.current,
        [key]: position,
      };
    },
    [],
  );

  const getScrollPosition = React.useCallback(
    (key: string) => scrollPositionsRef.current[key] || 0,
    [scrollPositionsRef],
  );

  const value = React.useMemo(
    () => ({
      saveScrollPosition,
      getScrollPosition,
    }),
    [saveScrollPosition, getScrollPosition],
  );

  return (
    <ScrollPositionContext.Provider value={ value }>
      {children}
    </ScrollPositionContext.Provider>
  );
}

/**
 * Hook that restores the position in the current URL path.
 * @param {React.RefObject<HTMLElement>} ref - the reference to the scrollable element
 * @param {boolean} viewLoaded - whether the view already exists in the DOM. Useful when content is loaded dynamically.
 */
export function useScrollRestoration(
  ref: React.RefObject<HTMLElement>,
  viewLoaded: boolean,
): void {
  const { saveScrollPosition, getScrollPosition } = useScrollPosition();
  const { pathname } = useLocation();

  React.useEffect(() => {
    if (ref.current) {
      const position = getScrollPosition(pathname || '');
      ref.current.scrollTo(0, position);
    }
  }, [pathname, ref, getScrollPosition]);

  React.useEffect(() => {
    function scrollHandler(): void {
      if (ref.current) {
        saveScrollPosition(pathname || '', ref.current.scrollTop);
      }
    }

    const element: HTMLElement = ref.current;
    if (element) {
      element.addEventListener('scroll', scrollHandler);
    }

    return () => {
      if (element) {
        element.removeEventListener('scroll', scrollHandler);
      }
    };
  }, [pathname, viewLoaded, ref, saveScrollPosition]);
}
