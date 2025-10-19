/* eslint-disable max-len */
import { useEffect } from 'react';
import { useLocation, useMatch } from 'react-router-dom';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';
import useStore from '@monetr/interface/hooks/store';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export default function MobileSidebar(): JSX.Element {
  const match = useMatch('/bank/:bankId/*');
  const isBankRoute = Boolean(match?.params?.bankId || null);
  const { setMobileSidebarOpen, mobileSidebarOpen } = useStore();
  const { pathname } = useLocation();

  // When we navigate away from the current page, if the sidebar is open; close it.
  // This achieves the behavior of; if they click a navigation item in the sidebar we automatically
  // close the sidebar. Without us having to have some kind of magic that listens for clicks or anything
  // like that.
  useEffect(() => {
    setMobileSidebarOpen(false);
  }, [setMobileSidebarOpen]);

  // Keeps the entire sidebar from re-rendering when things change. This way stuff like the drawer animation works
  // properly.
  const classNames = mergeTailwind(
    'z-40 w-screen h-screen top-0 left-0 bg-background flex flex-row backdrop-blur-sm dark:bg-opacity-50 backdrop-brightness-50',
    {
      fixed: mobileSidebarOpen,
      hidden: !mobileSidebarOpen,
    },
  );

  return (
    <div className={classNames}>
      <BankSidebar />
      {isBankRoute && <BudgetingSidebar className='w-auto flex-auto border-none overflow-y-auto' />}
    </div>
  );
}
