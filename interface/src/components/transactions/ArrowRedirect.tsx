import React from 'react';
import { NavigateFunction, useNavigate } from 'react-router-dom';
import { KeyboardArrowRight } from '@mui/icons-material';

export interface ArrowRedirectProps {
  redirect: string;
}

export default function ArrowRedirect({
  redirect,
}: ArrowRedirectProps): JSX.Element {
  const navigate: NavigateFunction = useNavigate();

  function go(): void {
    navigate(redirect);
  }

  return (
    <button className="flex-none dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis md:cursor-pointer" onClick={ go }>
      <KeyboardArrowRight />
    </button>
  );
}
