import { createContext } from 'react';

export interface IMobileSidebarContext {
  // The committed state; never changes while the drawer is mid motion, so consumers don't re-render every frame.
  // For the live position read the `data-sidebar-past` attribute on the document element from CSS.
  isOpen: boolean;
  open(): void;
  close(): void;
  toggle(): void;
}

// The provider lives in SwipeSidebar, which owns the motion and the DOM. This file only holds the shape so consumers
// like MSidebarToggle don't have to import the whole drawer.
export const MobileSidebarContext = createContext<IMobileSidebarContext>({
  isOpen: false,
  open: () => {},
  close: () => {},
  toggle: () => {},
});
