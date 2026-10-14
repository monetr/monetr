import { useContext } from 'react';
import { Link, useLocation } from 'wouter';

import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import Typography from '@monetr/interface/components/Typography';

import styles from './SettingsSidebarItem.module.scss';

export interface SettingsSidebarItemProps {
  icon: React.FC<{ className?: string }>;
  title: string;
  caption?: string;
  to: string;
  'data-testid'?: string;
}

export default function SettingsSidebarItem(props: SettingsSidebarItemProps): React.JSX.Element {
  const Icon = props.icon;
  const [pathname] = useLocation();
  const active = pathname === props.to;
  const { setIsOpen } = useContext(MobileSidebarContext);
  // Same as the budgeting sidebar, clicking the page we're already on just closes the mobile sidebar.
  const onClick = (event: React.MouseEvent) => {
    if (active) {
      event.preventDefault();
      setIsOpen(false);
    }
  };

  return (
    <Link
      className={styles.settingsSidebarItem}
      data-active={active}
      data-testid={props['data-testid']}
      onClick={onClick}
      to={props.to}
    >
      <Icon className={styles.settingsSidebarItemIcon} />
      <span className={styles.settingsSidebarItemText}>
        <Typography color='inherit' ellipsis size='lg' weight='medium'>
          {props.title}
        </Typography>
        {Boolean(props.caption) && (
          <Typography color='subtle' ellipsis size='xs'>
            {props.caption}
          </Typography>
        )}
      </span>
    </Link>
  );
}
