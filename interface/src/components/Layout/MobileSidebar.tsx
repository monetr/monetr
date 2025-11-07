import { useContext, useEffect } from 'react';
import { useLocation, useMatch } from 'react-router-dom';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';
import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export default function MobileSidebar(): JSX.Element {
  const { isOpen, setIsOpen } = useContext(MobileSidebarContext);
  const match = useMatch('/bank/:bankId/*');
  const isBankRoute = Boolean(match?.params?.bankId || null);
  const { pathname } = useLocation();

  // When we navigate away from the current page, if the sidebar is open; close it.
  // This achieves the behavior of; if they click a navigation item in the sidebar we automatically
  // close the sidebar. Without us having to have some kind of magic that listens for clicks or anything
  // like that.
  useEffect(() => {
    // Make sure that the pathname doesn't get autoremoved by the linter.
    if (pathname) {
      // Whenever the pathname changes close the sidebar
      setIsOpen(false);
    }
  }, [pathname, setIsOpen]);

  // Keeps the entire sidebar from re-rendering when things change. This way stuff like the drawer animation works
  // properly.
  const classNames = mergeTailwind(
    'z-40 w-screen h-screen top-0 left-0 bg-background flex flex-row backdrop-blur-sm dark:bg-opacity-50 backdrop-brightness-50',
    {
      fixed: isOpen,
      hidden: !isOpen,
    },
  );

  return (
    <div className={classNames}>
      <BankSidebar />
      {isBankRoute && <BudgetingSidebar className='w-auto flex-auto border-none overflow-y-auto' />}
    </div>
  );
}
