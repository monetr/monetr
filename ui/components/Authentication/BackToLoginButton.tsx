import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { ArrowBack } from '@mui/icons-material';
import { Button } from '@mui/material';

export default function BackToLoginButton(): JSX.Element {
  return (
    <div className="absolute top-2.5 left-2.5">
      <Button
        className="pr-3"
        component={ RouterLink }
        to="/login"
      >
        <ArrowBack className="mr-1" />
        Back To Login
      </Button>
    </div>
  );
}
