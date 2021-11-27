import { Backdrop, CircularProgress } from '@mui/material';
import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import useLogout from 'shared/authentication/actions/logout';

const Logout = (): JSX.Element => {
  const logout = useLogout();
  const navigate = useNavigate();

  useEffect(() => {
    logout().finally(() => navigate('/login'));
  })

  return (
    <Backdrop open={ true }>
      <CircularProgress color="inherit"/>
    </Backdrop>
  );
};

export default Logout;