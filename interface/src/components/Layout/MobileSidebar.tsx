import { useContext, useEffect } from 'react';
import { useLocation, useRoute } from 'wouter';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';
import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';

import styles from './MobileSidebar.module.scss';

export default function MobileSidebar(): JSX.Element {
  const { isOpen, setIsOpen } = useContext(MobileSidebarContext);
  const [, params] = useRoute<{ bankId: string }>('/bank/:bankId/*');
  const isBankRoute = Boolean(params?.bankId);
  const [pathname] = useLocation();

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

  return (
    <div className={styles.mobileSidebar} data-open={isOpen}>
      <BankSidebar />
      {isBankRoute && <BudgetingSidebar className={styles.budgetingSidebar} />}
    </div>
  );
}
