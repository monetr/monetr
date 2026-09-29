import { useCallback, useContext, useEffect, useMemo } from 'react';
import { CircleAlert, LogOut, Settings } from 'lucide-react';
import { Link } from 'wouter';

import Logo from '@monetr/interface/assets/Logo';
import Divider from '@monetr/interface/components/Divider';
import BankSidebarItem from '@monetr/interface/components/Layout/BankSidebarItem';
import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import MSidebarToggle from '@monetr/interface/components/MSidebarToggle';
import Typography from '@monetr/interface/components/Typography';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { useLinks } from '@monetr/interface/hooks/useLinks';
import { usePatchUser } from '@monetr/interface/hooks/usePatchUser';
import { preloadSortableList, useSortableList } from '@monetr/interface/hooks/useSortableList';
import type { ID } from '@monetr/interface/models/ID';
import type MonetrLink from '@monetr/interface/models/Link';
import mergeClasses from '@monetr/interface/util/mergeClasses';
import { useSnackbar } from '@monetr/notify';

import BankSidebarSubscriptionItem from './BankSidebarSubscriptionItem';

import styles from './BankSidebar.module.scss';

export interface BankSidebarProps {
  className?: string;
}

export default function BankSidebar(props: BankSidebarProps): React.JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  const { data: links, isLoading, isError } = useLinks();
  const { data: authentication } = useAuthentication();
  const user = authentication?.user;
  const patchUser = usePatchUser();
  const { enqueueSnackbar } = useSnackbar();

  // Alphabetical is the default order, anything the user hasn't placed yet (like a newly added link) lands here. Copy
  // before sorting, sorting in place would reorder the links query cache out from under everything else.
  const linksSorted = useMemo(
    () =>
      [...(links ?? [])].sort((a, b) => {
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
      }),
    [links],
  );

  const onReorder = useCallback(
    (linkOrder: Array<ID<MonetrLink>>) => {
      if (!user) {
        return;
      }
      // The patch updates the current user in the cache right away, and puts it back if this fails.
      patchUser({ userId: user.userId, linkOrder }).catch(
        () =>
          void enqueueSnackbar('Failed to save the order of your links.', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      );
    },
    [user, patchUser, enqueueSnackbar],
  );

  // On desktop the drag code loads when you hover a link, but a touch has no hover before it and the long press to
  // start a drag comes right after the tap. So on mobile fetch it as soon as the sidebar opens instead. Only the mobile
  // toggle ever opens it, so this never fires on desktop.
  const { isOpen: isMobileSidebarOpen } = useContext(MobileSidebarContext);
  useEffect(() => {
    if (isMobileSidebarOpen) {
      preloadSortableList();
    }
  }, [isMobileSidebarOpen]);

  const sortable = useSortableList({
    items: linksSorted,
    order: user?.linkOrder,
    getId: getLinkId,
    onReorder,
  });

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

  // TODO Make it so that when we are in the "add link" page, we have the add link +1 button as active.
  return (
    <SidebarWrapper className={props.className}>
      <div className={styles.sortableList} ref={sortable.containerRef}>
        {sortable.items.map(link => (
          <BankSidebarItem key={link.linkId} link={link} sortable={sortable.getItemProps(link.linkId)} />
        ))}
      </div>
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

function getLinkId(link: MonetrLink): ID<MonetrLink> {
  return link.linkId;
}

interface SidebarWrapperProps {
  className?: string;
  children?: React.ReactNode;
}

function SidebarWrapper(props: SidebarWrapperProps): React.JSX.Element {
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

function SettingsButton(): React.JSX.Element {
  return (
    <Link data-testid='bank-sidebar-settings' to='/settings'>
      <Settings className={styles.actionIcon} />
    </Link>
  );
}

function LogoutButton(): React.JSX.Element {
  // By doing reloadDocument, we are forcing the @tanstack/react-query cache to be emptied. This will naturally just
  // make it easier to prevent the current user's data from leaking into another session.
  return (
    <Link data-testid='bank-sidebar-logout' to='/logout'>
      <LogOut className={styles.actionIcon} />
    </Link>
  );
}
