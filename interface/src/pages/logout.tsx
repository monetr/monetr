import type React from 'react';

import MSpinner from '@monetr/interface/components/MSpinner';
import useLogout from '@monetr/interface/hooks/useLogout';
import useMountEffect from '@monetr/interface/hooks/useMountEffect';

export default function LogoutPage(): React.JSX.Element {
  const logout = useLogout();
  useMountEffect(() => {
    logout().finally(() => window.location.replace('/login'));
  });

  return (
    <div className='flex flex-col justify-center items-center'>
      <MSpinner />
    </div>
  );
}
