import { Typography } from '@mui/material';
import React, { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { getRelease } from 'shared/bootstrap/selectors';

const GlobalFooter = (): JSX.Element => {
  return (
    <Typography
      className="absolute inline w-full text-center bottom-1 opacity-30"
    >
      © { new Date().getFullYear() } monetr LLC
    </Typography>
  );
};

export default GlobalFooter;
