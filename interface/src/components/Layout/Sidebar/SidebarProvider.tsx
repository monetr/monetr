import { createContext, useState } from 'react';

import useIsMobile from '@monetr/interface/hooks/useIsMobile';

const DEFAULT_SIDEBAR_WIDTH = 22;

export interface ISidebarContext {
  isOpen: boolean;
  // Width represents the width of the sidebar in rem units.
  width: number;
  setIsOpen: (_: boolean) => void;
  setWidth: (_: number) => void;
}

export const SidebarContext = createContext<ISidebarContext>({
  isOpen: false,
  width: DEFAULT_SIDEBAR_WIDTH,
  setIsOpen: (_: boolean) => {},
  setWidth: (_: number) => {},
});

export interface SidebarProviderProps {
  children: React.ReactNode;
}

export default function SidebarProvider(props: SidebarProviderProps): React.JSX.Element {
  const isMobile = useIsMobile();
  const [isOpen, setIsOpen] = useState(!isMobile);
  const [width, setWidth] = useState(DEFAULT_SIDEBAR_WIDTH);
  return (
    <SidebarContext.Provider value={{ width, isOpen, setIsOpen, setWidth }}>
      <div
        // Pass our width as a variable, so that way every component rendered within the sidebar context is aware of how
        // wide the sidebar is. This way the child components can also use CSS properties to pad themselves
        // appropriately.
        style={
          {
            '--sidebar-width': `${width}rem`,
            // Add padding to our sidebar provider on the left side IF the sidebar is open and we are not on mobile. On
            // mobile the sidebar is rendered _over_ the content of the page and should not affect the padding.
            paddingLeft: isOpen && !isMobile ? `${width}rem` : 'unset',
            width: '100%',
            display: 'flex',
            flexDirection: 'row',
          } as React.CSSProperties
        }
      >
        {props.children}
      </div>
    </SidebarContext.Provider>
  );
}
