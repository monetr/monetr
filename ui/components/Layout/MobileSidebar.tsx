/* eslint-disable max-len */
import React, { useEffect } from 'react';
import { useLocation, useMatch } from 'react-router-dom';

import BankSidebar from 'components/Layout/BankSidebar';
import BudgetingSidebar from 'components/Layout/BudgetingSidebar';
import useStore from 'hooks/store';

export default function MobileSidebar(): JSX.Element {
  const match = useMatch('/bank/:bankId/*');
  const isBankRoute = Boolean(+match?.params?.bankId || null);
  const { setMobileSidebarOpen, mobileSidebarOpen } = useStore();
  const { pathname } = useLocation();

  // When we navigate away from the current page, if the sidebar is open; close it.
  // This achieves the behavior of; if they click a navigation item in the sidebar we automatically
  // close the sidebar. Without us having to have some kind of magic that listens for clicks or anything
  // like that.
  useEffect(() => {
    setMobileSidebarOpen(false);
  }, [pathname, setMobileSidebarOpen]);

  if (!mobileSidebarOpen) {
    return null;
  }

  return (
    <div className='fixed z-40 w-screen h-screen top-0 left-0 lg:hidden dark:bg-dark-monetr-background flex flex-row backdrop-blur-sm dark:bg-opacity-50 backdrop-brightness-50'>
      <BankSidebar />
      { isBankRoute && <BudgetingSidebar className='w-auto flex-grow border-none overflow-y-auto' /> }
    </div>
  );
}
