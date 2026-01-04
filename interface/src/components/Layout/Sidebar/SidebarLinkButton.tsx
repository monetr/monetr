import { Link } from 'react-router-dom';

import styles from './SidebarLinkButton.module.scss';

export interface SidebarLinkButtonProps {
  to: string;
  reloadDocument?: boolean;
  icon: React.FC<{ className?: string }>;
}

export default function SidebarLinkButton(props: SidebarLinkButtonProps): React.JSX.Element {
  const Icon = props.icon;
  return (
    <Link reloadDocument={props.reloadDocument} to={props.to}>
      <Icon className={styles.sidebarLinkButton} />
    </Link>
  );
}
