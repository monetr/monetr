import { useCallback, useEffect, useRef } from 'react';

const FOCUSABLE = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

export interface DrawerFocusOptions {
  isMobile: boolean;
  isOpen: boolean;
  panelRef: React.RefObject<HTMLElement | null>;
  close(): void;
}

export interface DrawerFocus {
  // Record the currently focused element so it can be restored when the drawer closes. This has to run before React
  // commits `inert` onto the page, which blurs the focused element before any effect gets a look at it.
  rememberFocus(): void;
}

// Scroll lock, focus trap and keyboard handling for the open drawer.
export default function useDrawerFocus(options: DrawerFocusOptions): DrawerFocus {
  const { isMobile, isOpen, panelRef, close } = options;

  const restoreFocusRef = useRef<HTMLElement | null>(null);

  const rememberFocus = useCallback(() => {
    restoreFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  }, []);

  // Scroll lock and focus hand off.
  useEffect(() => {
    if (!isMobile) {
      return;
    }
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      panelRef.current?.focus({ preventScroll: true });
      return;
    }

    document.body.style.overflow = '';
    const restore = restoreFocusRef.current;
    restoreFocusRef.current = null;
    if (restore && document.contains(restore)) {
      restore.focus({ preventScroll: true });
    }
  }, [isMobile, isOpen, panelRef]);

  useEffect(() => {
    if (!isMobile || !isOpen) {
      return;
    }

    const onKeyDown = (event: KeyboardEvent) => {
      const panel = panelRef.current;
      // Dropdowns, bottom drawers and modals opened from inside the panel portal outside it (Radix and vaul to the
      // body, NiceModal to the app root) and own the keyboard while they are up.
      if (!panel?.contains(event.target as Node)) {
        return;
      }

      if (event.key === 'Escape') {
        close();
        return;
      }
      if (event.key !== 'Tab') {
        return;
      }

      const focusable = panel.querySelectorAll<HTMLElement>(FOCUSABLE);
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (!first || !last) {
        return;
      }
      if (event.shiftKey && document.activeElement === first) {
        last.focus();
        event.preventDefault();
        return;
      }
      if (!event.shiftKey && document.activeElement === last) {
        first.focus();
        event.preventDefault();
      }
    };

    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [isMobile, isOpen, close, panelRef]);

  return { rememberFocus };
}
