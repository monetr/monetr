import { Button } from '@mui/material';
import classnames from 'classnames';
import React from 'react';
import { Link as RouterLink, useLocation } from 'react-router-dom';

interface SidebarButtonProps {
  onClick?: () => void;
  children: React.ReactNode;
  to: string;
}

export default function SidebarButton(props: SidebarButtonProps): JSX.Element {
  const location = useLocation();
  let className = 'justify-start text-lg w-full';

  return (
    <div className={ classnames(className, 'sidebar-button-wrapper', {
      'sidebar-button-wrapper-active': location.pathname === props.to,
    }) }>
      <div className="navigation-before"/>
      <Button
        onClick={ props.onClick }
        to={ props.to }
        className={ className }
        component={ RouterLink }
        children={ props.children }
        color="inherit"
      />
      <div className="navigation-after"/>
    </div>
  )

}
