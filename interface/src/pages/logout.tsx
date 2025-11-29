import type React from 'react';

import { flexVariants } from '@monetr/interface/components/Flex';
import { heights, widths } from '@monetr/interface/components/Layout';
import MSpinner from '@monetr/interface/components/MSpinner';
import useLogout from '@monetr/interface/hooks/useLogout';
import useMountEffect from '@monetr/interface/hooks/useMountEffect';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export default function LogoutPage(): React.JSX.Element {
  const logout = useLogout();
  useMountEffect(() => {
    logout().finally(() => window.location.replace('/login'));
  });

  return (
    <div
      className={mergeTailwind(
        flexVariants({ justify: 'center', align: 'center', orientation: 'column' }),
        heights.screen,
        widths.screen,
      )}
    >
      <MSpinner />
    </div>
  );
}
