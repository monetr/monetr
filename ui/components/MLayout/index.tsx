import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowBackOutlined, Menu } from '@mui/icons-material';

import MSidebar from 'components/MSidebar';
import { ReactElement } from 'components/types';

export interface MLayoutProps {
  children?: ReactElement;
}

export default function MLayout(props: MLayoutProps): JSX.Element {
  const navigate = useNavigate();
  const [isSidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="h-full flex min-h-full min-w-0">
      <MSidebar open={ isSidebarOpen } setClosed={ () => setSidebarOpen(false) } />
      <div className="relative lg:ml-64 flex w-0 min-w-0 flex-1 flex-col">
        <div className="fixed left-0 right-0 top-0 z-10 flex flex-shrink-0 bg-purple-800 lg:left-64">
          <div className="flex flex-1 items-center justify-between px-4 md:pr-4 md:pl-4 h-16">
            <button
              className="mr-2 text-white focus:outline-none lg:hidden"
              onClick={ () => setSidebarOpen(true) }
            >
              <Menu />
            </button>
            <button
              className="mr-2 text-white focus:outline-none"
              onClick={ () => navigate(-1) }
            >
              <ArrowBackOutlined />
            </button>
            <div className="flex flex-1">
            </div>
            <div className="flex items-center">
              <span className="text-white">bank</span>
            </div>
          </div>
        </div>
        <div className="m-content">
          <div className="m-view-area">
            { props.children }
          </div>
        </div>
      </div>
    </div>
  );
}
