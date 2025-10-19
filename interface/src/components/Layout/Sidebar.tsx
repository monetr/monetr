import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import MobileSidebar from '@monetr/interface/components/Layout/MobileSidebar';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';

export default function Sidebar(): JSX.Element {
  const isMobile = useIsMobile();
  if (isMobile) {
    return <MobileSidebar />;
  }

  return <BankSidebar />;
}
