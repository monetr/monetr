import { createContext, useState } from 'react';

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
  return <MobileSidebarContext.Provider value={{ isOpen, setIsOpen }}>{props.children}</MobileSidebarContext.Provider>;
}
