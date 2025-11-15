import { createContext, useEffect, useState } from 'react';

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
  const [isOpen, setIsOpen] = useState(false);
  useEffect(() => {
    const root = document.querySelector('#root');
    if (root) {
      root.className = `relative ${isOpen ? 'sidebar-open' : 'sidebar-close'}`;
    }
  });
  return <MobileSidebarContext.Provider value={{ isOpen, setIsOpen }}>{props.children}</MobileSidebarContext.Provider>;
}
