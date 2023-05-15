import React from 'react';
import { Link, useLocation } from 'react-router-dom';

import { ReactElement } from './types';

import clsx from 'clsx';

export interface MSidebarButton {
  children: ReactElement;
  to: string;
}

export default function MSidebarButton(props: MSidebarButton): JSX.Element {
  const location = useLocation();
  const active = location.pathname === props.to;

  const className = clsx(
    'flex',
    'gap-x-3',
    'rounded-lg',
    'p-2',
    'text-sm',
    'font-semibold',
    'leading-6',
    'text-gray-50',
    {
      'bg-purple-700': active,
      'hover:bg-purple-700': !active,
    },
  );

  return (
    <li>
      <Link { ...props } className={ className } />
    </li>
  );
}
