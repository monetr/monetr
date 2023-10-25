import React from 'react';
import { Backdrop, CircularProgress } from '@mui/material';

import useLogout from 'hooks/useLogout';
import useMountEffect from 'hooks/useMountEffect';

export default function LogoutPage(): JSX.Element {
  const logout = useLogout();
  useMountEffect(() => {
    logout().finally(() => window.location.replace('/login'));
  });

  return (
    <Backdrop open={ true }>
      <CircularProgress color="inherit" />
    </Backdrop>
  );
}
