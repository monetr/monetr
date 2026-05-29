import { CircleAlert, LogOut, Settings } from 'lucide-react';
import { Link } from 'wouter';

import Logo from '@monetr/interface/assets/Logo';
import Divider from '@monetr/interface/components/Divider';
import BankSidebarItem from '@monetr/interface/components/Layout/BankSidebarItem';
import MSidebarToggle from '@monetr/interface/components/MSidebarToggle';
import Typography from '@monetr/interface/components/Typography';
import { useLinks } from '@monetr/interface/hooks/useLinks';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './BankSidebar.module.scss';
import BankSidebarSubscriptionItem from './BankSidebarSubscriptionItem';

export interface BankSidebarProps {
  className?: string;
}

export default function BankSidebar(props: BankSidebarProps): JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  const { data: links, isLoading, isError } = useLinks();
  if (isLoading) {
    return <SidebarWrapper className={props.className} />;
  }

  if (isError) {
    return (
      <SidebarWrapper className={props.className}>
        <div className={styles.itemRow}>
          <div className={styles.indicatorCircle}>
            <CircleAlert className={styles.alertIcon} />
          </div>
        </div>
      </SidebarWrapper>
    );
  }

  const linksSorted = (links ?? []).sort((a, b) => {
    const nameA = a.getName().toUpperCase();
    const nameB = b.getName().toUpperCase();
    if (nameA < nameB) {
      return -1;
    }
    if (nameA > nameB) {
      return 1;
    }

    // names must be equal
    return 0;
  });

  // TODO Make it so that when we are in the "add link" page, we have the add link +1 button as active.
  return (
    <SidebarWrapper className={props.className}>
      {linksSorted.map(link => (
        <BankSidebarItem key={link.linkId} link={link} />
      ))}
      <div className={styles.itemRow}>
        <Link className={styles.indicatorCircle} to='/link/create'>
          <Typography color='emphasis' size='xl' weight='bold'>
            +1
          </Typography>
        </Link>
      </div>
    </SidebarWrapper>
  );
}

interface SidebarWrapperProps {
  className?: string;
  children?: React.ReactNode;
}

function SidebarWrapper(props: SidebarWrapperProps): JSX.Element {
  return (
    <div className={mergeClasses(styles.bankSidebarWrapperRoot, props.className)} data-testid='bank-sidebar'>
      <MSidebarToggle className={styles.toggle} />
      <div className={styles.logoWrapper}>
        <Logo className={styles.logo} />
      </div>
      <Divider className={styles.divider} />
      <div className={styles.items}>{props?.children}</div>
      <BankSidebarSubscriptionItem />
      <SettingsButton />
      <LogoutButton />
    </div>
  );
}

function SettingsButton(): JSX.Element {
  return (
    <Link data-testid='bank-sidebar-settings' to='/settings'>
      <Settings className={styles.actionIcon} />
    </Link>
  );
}

function LogoutButton(): JSX.Element {
  // By doing reloadDocument, we are forcing the @tanstack/react-query cache to be emptied. This will naturally just make it
  // easier to prevent the current user's data from leaking into another session.
  return (
    <Link data-testid='bank-sidebar-logout' to='/logout'>
      <LogOut className={styles.actionIcon} />
    </Link>
  );
}
