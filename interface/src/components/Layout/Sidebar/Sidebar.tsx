import { LogOut, Settings } from 'lucide-react';

import BankSidebarSubscriptionItem from '@monetr/interface/components/Layout/BankSidebarSubscriptionItem';
import SidebarLinkButton from '@monetr/interface/components/Layout/Sidebar/SidebarLinkButton';
import { Sheet } from '@monetr/interface/components/Sheet';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import { useContext } from 'react';
import { SidebarContext } from '@monetr/interface/components/Layout/Sidebar/SidebarProvider';
import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';

import styles from './Sidebar.module.scss';

export default function Sidebar(): React.JSX.Element {
  const { isOpen, setIsOpen } = useContext(SidebarContext);
  const isMobile = useIsMobile();

  if (isMobile) {
    return (
      <Sheet open={isOpen}>
        <BankSidebar />
      </Sheet>
    );
  }

  return (
    <div className={styles.sidebarRoot}>
      <BankSidebarSubscriptionItem />
      <SidebarLinkButton icon={Settings} to='/settings' />
      <SidebarLinkButton icon={LogOut} reloadDocument to='/logout' />
    </div>
  );
}
