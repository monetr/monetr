import React from 'react';

import MSidebarToggle from './MSidebarToggle';
import { ReactElement } from './types';

export interface MTopNavigationProps {
  icon: React.FC;
  title: string;
  children?: ReactElement;
}

export default function MTopNavigation(props: MTopNavigationProps): JSX.Element {
  const Icon = props.icon;
  return (
    <div className='w-full h-12 flex-none flex items-center px-4 gap-4 justify-between'>
      <div className='flex gap-4'>
        <MSidebarToggle />
        <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
          <Icon />
          { props.title }
        </span>
      </div>
      { props.children }
    </div>
  );
}
