import { Box, styled } from "@mui/material";
import React from 'react';

export default function VerticalPuller(): JSX.Element {
  const Thing = styled(Box)(({ theme }) => ({
    width: 6,
    height: 150,
    backgroundColor: theme.palette.mode === 'light' ? theme.palette.grey[300] : theme.palette.grey[900],
    borderRadius: 3,
    position: 'absolute',
    top: 'calc(50% - 75px)',
    left: 9,
  }));

  return <Thing />
}

