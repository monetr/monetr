import { createContext, useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

export interface IMobileSidebarContext {
  isOpen: boolean;
  setIsOpen: (_: boolean) => void;
}

export const MobileSidebarContext = createContext<IMobileSidebarContext>({
  isOpen: false,
  setIsOpen: (_: boolean) => {},
});

export interface MobileSidebarContextProviderProps {
  children: React.ReactNode;
}

export default function MobileSidebarContextProvider(props: MobileSidebarContextProviderProps): JSX.Element {
  const { pathname } = useLocation();
  const [isOpen, setIsOpen] = useState(false);
  useEffect(() => {
    const root = document.querySelector('#root');
    if (root) {
      root.className = `${isOpen ? 'sidebar-open' : 'sidebar-close'}`;
    }
  });

  // When we navigate away from the current page, if the sidebar is open; close it.
  // This achieves the behavior of; if they click a navigation item in the sidebar we automatically
  // close the sidebar. Without us having to have some kind of magic that listens for clicks or anything
  // like that.
  useEffect(() => {
    // Make sure that the pathname doesn't get autoremoved by the linter.
    if (pathname) {
      // Whenever the pathname changes close the sidebar
      setIsOpen(false);
    }
  }, [pathname]);

  return <MobileSidebarContext.Provider value={{ isOpen, setIsOpen }}>{props.children}</MobileSidebarContext.Provider>;
}
