import React, { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import MLogo from 'components/MLogo';
import MSpan from 'components/MSpan';
import request from 'util/request';


export default function VerifyEmail(): JSX.Element {
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

  return <VerifyEmailView />;
}

export function VerifyEmailView(): JSX.Element {
  return (
    <div className='w-full h-full flex flex-col justify-center items-center gap-2 p-4'>
      <MLogo className='h-24 w-24' />
      <MSpan size='2xl' weight='bold'>
        Email Verification
      </MSpan>
      <MSpan size='xl' className='text-center'>
        Your email is being verified, one moment...
      </MSpan>
    </div>
  );
}
