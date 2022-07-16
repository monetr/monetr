import React from 'react';
import { Typography } from '@mui/material';

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
