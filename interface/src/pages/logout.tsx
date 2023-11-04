import React from 'react';
import { Backdrop, CircularProgress } from '@mui/material';

import useLogout from '@monetr/interface/hooks/useLogout';
import useMountEffect from '@monetr/interface/hooks/useMountEffect';

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
