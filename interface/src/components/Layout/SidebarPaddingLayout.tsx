import { Outlet } from 'react-router-dom';

import styles from './SidebarPaddingLayout.module.scss';

export default function SidebarPaddingLayout(): React.JSX.Element {
  return (
    <div className={styles.sidebarPaddingLayout}>
      <Outlet />
    </div>
  );
}
