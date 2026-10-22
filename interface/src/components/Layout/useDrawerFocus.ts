import { useCallback, useEffect, useRef } from 'react';

const focusableSelector = [
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
  // rememberFocus needs to be called before React sets inert on the page. Inert blurs whatever has focus before any of
  // our effects run, so by then we wouldn't know what to give focus back to when the drawer closes.
  rememberFocus(): void;
}

export default function useDrawerFocus(options: DrawerFocusOptions): DrawerFocus {
  const { isMobile, isOpen, panelRef, close } = options;

  const restoreFocusRef = useRef<HTMLElement | null>(null);

  const rememberFocus = useCallback(() => {
    restoreFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  }, []);

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
      // If the key press came from a dropdown or modal that was opened from the panel then leave it alone. Those portal
      // outside of the panel and handle their own keyboard.
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

      const focusable = panel.querySelectorAll<HTMLElement>(focusableSelector);
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
