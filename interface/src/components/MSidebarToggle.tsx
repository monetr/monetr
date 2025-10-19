import { ArrowLeft, PanelLeft, PanelLeftClose } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

import useStore from '@monetr/interface/hooks/store';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MSidebarToggleProps {
  backButton?: string;
  className?: string;
}

export default function MSidebarToggle(props: MSidebarToggleProps): JSX.Element {
  const navigate = useNavigate();
  const { setMobileSidebarOpen, mobileSidebarOpen } = useStore();

  function onClick() {
    if (props.backButton) {
      navigate(props.backButton);
    } else {
      setMobileSidebarOpen(!mobileSidebarOpen);
    }
  }

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
      {!mobileSidebarOpen && <PanelLeft />}
      {mobileSidebarOpen && <PanelLeftClose />}
    </button>
  );
}
