import React from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowBackOutlined, MenuOpenOutlined, MenuOutlined } from '@mui/icons-material';

import useStore from '@monetr/interface/hooks/store';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MSidebarToggleProps {
  backButton?: string;
  className?: string;
}

export default function MSidebarToggle(props: MSidebarToggleProps): JSX.Element {
  const navigate = useNavigate();
  const { setMobileSidebarOpen, mobileSidebarOpen } = useStore();

  function onClick() {
    if (!!props.backButton) {
      navigate(props.backButton);
    } else {
      setMobileSidebarOpen(!mobileSidebarOpen);
    }
  }

  const className = mergeTailwind(
    'visible lg:hidden',
    'dark:text-dark-monetr-content-emphasis cursor-pointer h-12 flex items-center justify-center',
    props.className,
  );

  if (props.backButton) {
    return (
      <div className={ className } onClick={ onClick }>
        <ArrowBackOutlined />
      </div>
    );
  }

  return (
    <div className={ className } onClick={ onClick }>
      { !mobileSidebarOpen && <MenuOutlined /> }
      { mobileSidebarOpen && <MenuOpenOutlined /> }
    </div>
  );
}
