import { useCallback, useContext } from 'react';
import { ArrowLeft, PanelLeft, PanelLeftClose } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MSidebarToggleProps {
  backButton?: string;
  className?: string;
}

export default function MSidebarToggle(props: MSidebarToggleProps): JSX.Element {
  const navigate = useNavigate();
  const { isOpen, setIsOpen } = useContext(MobileSidebarContext);

  const onClick = useCallback(() => {
    if (props.backButton) {
      navigate(props.backButton);
    } else {
      setIsOpen(!isOpen);
    }
  }, [props.backButton, isOpen, navigate, setIsOpen]);

  const className = mergeTailwind(
    'visible lg:hidden',
    'dark:text-dark-monetr-content-emphasis h-12 flex items-center justify-center',
    props.className,
  );

  if (props.backButton) {
    return (
      <button type='button' className={className} onClick={onClick}>
        <ArrowLeft />
      </button>
    );
  }

  return (
    <button type='button' className={className} onClick={onClick}>
      {!isOpen && <PanelLeft />}
      {isOpen && <PanelLeftClose />}
    </button>
  );
}
