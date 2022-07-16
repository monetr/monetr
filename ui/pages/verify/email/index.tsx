import React, { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { CircularProgress, Typography } from '@mui/material';

import CenteredLogo from 'components/Logo/CenteredLogo';
import request from 'util/request';

export default function VerifyEmailPage(): JSX.Element {
  const location = useLocation();
  const navigate = useNavigate();

  function errorRedirect(message: string, nextUrl: string = '/login') {
    window.alert(message);
    navigate(nextUrl);
  }

  const search = location.search;
  const query = new URLSearchParams(search);
  const token = query.get('token');

  useEffect(() => {
    if (!token) {
      errorRedirect('Email verification link is not valid.');
      return;
    }

    request().post('/authentication/verify', {
      'token': token,
    })
      .then(result => errorRedirect(
        result?.data?.message || 'Your email has been verified, please login.',
        result?.data?.nextUrl || '/login',
      ))
      .catch(error => errorRedirect(
        error?.response?.data?.error || 'Failed to verify email address.',
        error?.response?.data?.nextUrl,
      ));
  }, [token]);

  return (
    <div className="flex items-center justify-center w-full h-full max-h-full">
      <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
        <CenteredLogo />
        <div className="w-full pt-2.5 pb-2.5">
          <Typography
            variant="h5"
            className="w-full text-center"
          >
            Verifying email address...
          </Typography>
        </div>
        <div className="w-full pt-2.5 pb-2.5 flex justify-center">
          <CircularProgress />
        </div>
      </div>
    </div>
  );
};
