import { ButtonBase, ButtonBaseProps } from '@mui/material';
import clsx from 'clsx';
import React from 'react';

export interface MButtonProps extends ButtonBaseProps {

}

export default function MButton(props: MButtonProps): JSX.Element {
  const classNames = clsx(
    'bg-purple-700',
    'dark:bg-purple-600',
    'dark:focus:ring-purple-800',
    'dark:hover:bg-purple-700',
    'focus:outline-none',
    'font-medium',
    'hover:bg-purple-800',
    'mb-2',
    'mr-2',
    'px-5',
    'py-2.5',
    'rounded-lg',
    'text-sm',
    'text-white',
    'w-full',
  )

  return <ButtonBase
    { ...props }
    className={ classNames }
  />;
}
