import { useCallback, useContext } from 'react';
import { ArrowLeft, PanelLeft, PanelLeftClose } from 'lucide-react';
import { useLocation } from 'wouter';

import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './MSidebarToggle.module.scss';

export interface MSidebarToggleProps {
  backButton?: string;
  className?: string;
}

export default function MSidebarToggle(props: MSidebarToggleProps): React.JSX.Element {
  const [, navigate] = useLocation();
  const { isOpen, setIsOpen } = useContext(MobileSidebarContext);

  const onClick = useCallback(() => {
    // Same as the title in MTopNavigation, if we came from the page this points to then actually go back to it.
    if (props.backButton && window.history.state?.from === props.backButton) {
      window.history.back();
    } else if (props.backButton) {
      navigate(props.backButton);
    } else {
      setIsOpen(!isOpen);
    }
  }, [props.backButton, isOpen, navigate, setIsOpen]);

  const className = mergeClasses(styles.toggle, props.className);

  if (props.backButton) {
    return (
      <button className={className} onClick={onClick} type='button'>
        <ArrowLeft />
      </button>
    );
  }

  return (
    <button className={className} onClick={onClick} type='button'>
      {!isOpen && <PanelLeft />}
      {isOpen && <PanelLeftClose />}
    </button>
  );
}
