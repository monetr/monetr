import React from 'react';
import { Link as RouterLink, useLocation } from 'react-router-dom';
import { Button } from '@mui/material';
import clsx from 'clsx';

interface SidebarButtonProps {
  onClick?: () => void;
  children: React.ReactNode;
  to: string;
  prefix?: string;
}

export default function SidebarButton(props: SidebarButtonProps): JSX.Element {
  const location = useLocation();
  const className = 'justify-start text-lg w-full';
  const selected = location.pathname === props.to ||
    (props.prefix && location.pathname.startsWith(props.prefix));

  return (
    <div className={ clsx(className, 'sidebar-button-wrapper', {
      'sidebar-button-wrapper-active': selected,
    }) }>
      <div className="navigation-before" />
      <Button
        onClick={ props.onClick }
        to={ props.to }
        className={ className }
        component={ RouterLink }
        children={ props.children }
        color="inherit"
      />
      <div className="navigation-after" />
    </div>
  );

}
