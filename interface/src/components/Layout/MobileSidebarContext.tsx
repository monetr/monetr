import { createContext } from 'react';

export interface IMobileSidebarContext {
  // isOpen only changes once the drawer commits to opening or closing, not while it is moving. If you need the live
  // position then use the data-sidebar-past attribute on the html element from CSS.
  isOpen: boolean;
  open(): void;
  close(): void;
  toggle(): void;
}

export const MobileSidebarContext = createContext<IMobileSidebarContext>({
  isOpen: false,
  open: () => {},
  close: () => {},
  toggle: () => {},
});
