import React from 'react';
import { MenuOpenOutlined, MenuOutlined } from '@mui/icons-material';

import useStore from 'hooks/store';
import mergeTailwind from 'util/mergeTailwind';

export interface MSidebarToggleProps {
  className?: string;
}

export default function MSidebarToggle(props: MSidebarToggleProps): JSX.Element {
  const { setMobileSidebarOpen, mobileSidebarOpen } = useStore();

  function onClick() {
    setMobileSidebarOpen(!mobileSidebarOpen);
  }

  const className = mergeTailwind(
    'visible lg:hidden',
    'dark:text-dark-monetr-content-emphasis cursor-pointer h-12 flex items-center justify-center',
    props.className,
  );

  return (
    <div className={ className } onClick={ onClick }>
      { !mobileSidebarOpen && <MenuOutlined /> }
      { mobileSidebarOpen && <MenuOpenOutlined /> }
    </div>
  );
}
