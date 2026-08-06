import { useCallback, useContext } from 'react';
import { ArrowLeft, PanelLeft, PanelLeftClose } from 'lucide-react';
import { useLocation } from 'wouter';

import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContext';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './MSidebarToggle.module.scss';

export interface MSidebarToggleProps {
  backButton?: string;
  className?: string;
}

export default function MSidebarToggle(props: MSidebarToggleProps): React.JSX.Element {
  const [, navigate] = useLocation();
  const { toggle } = useContext(MobileSidebarContext);

  const onClick = useCallback(() => {
    // Same as the title in MTopNavigation, if we came from the page this points to then actually go back to it.
    if (props.backButton && window.history.state?.from === props.backButton) {
      window.history.back();
    } else if (props.backButton) {
      navigate(props.backButton);
      return;
    }
    toggle();
  }, [props.backButton, navigate, toggle]);

  const className = mergeClasses(styles.toggle, props.className);

  if (props.backButton) {
    return (
      <button className={className} onClick={onClick} type='button'>
        <ArrowLeft />
      </button>
    );
  }

  // Both glyphs render and CSS swaps them off the drawer's live position, so the icon stays right mid drag without a
  // re-render on the frame path.
  return (
    <button className={className} onClick={onClick} type='button'>
      <PanelLeft className={styles.iconClosed} />
      <PanelLeftClose className={styles.iconOpen} />
    </button>
  );
}
