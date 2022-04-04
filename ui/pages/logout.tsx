import { Backdrop, CircularProgress } from '@mui/material';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import useLogout from 'shared/authentication/actions/logout';
import useMountEffect from 'shared/util/useMountEffect';

export default function LogoutPage(): JSX.Element {
  const logout = useLogout();
  const navigate = useNavigate();
  useMountEffect(() => {
    logout().finally(() => navigate('/login'));
  })

  return (
    <Backdrop open={ true }>
      <CircularProgress color="inherit"/>
    </Backdrop>
  );
}
